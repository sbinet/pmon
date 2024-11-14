// Copyright 2015 The pmon Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/sbinet/pmon"
)

var (
	freq = flag.Duration("freq", 1*time.Second, "frequence to capture resource usage")
	out  = flag.String("o", "pmon.data", "path to file to store resources usage log")
	pid  = flag.Int("p", 0, "PID of an already running process to monitor")

	usage = `pmon monitors process resources usage.

Usage:

 $ pmon [options] command [command-arg1 [command-arg2 [...]]]

Example:

 $ pmon my-command arg0 arg1
 $ pmon -- my-command arg0 arg1
 $ pmon -p 1234

Options:
`
)

func main() {

	log.SetFlags(0)
	log.SetPrefix("pmon: ")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "%s", usage)
		flag.PrintDefaults()
	}

	flag.Parse()

	if *pid <= 0 && flag.NArg() <= 0 {
		log.Printf("expect a command (and its arguments) as argument")
		flag.Usage()
		os.Exit(1)
	}

	err := os.MkdirAll(filepath.Dir(*out), 0755)
	if err != nil {
		log.Fatalf("could not create output directory: %+v", err)
	}

	switch {
	case *pid > 0:
		runPID(*out, *pid)
	default:
		cmd := flag.Arg(0)
		args := flag.Args()[1:]
		runCmd(*out, cmd, args)
	}
}

func runCmd(out, cmd string, args []string) {
	f, err := os.Create(out)
	if err != nil {
		log.Fatalf("could not create output log file: %+v", err)
	}
	defer f.Close()

	w := bufio.NewWriter(f)

	proc := pmon.New(cmd, args...)
	proc.W = w
	proc.Freq = *freq

	go func() {
		sigch := make(chan os.Signal, 1)
		signal.Notify(sigch, os.Interrupt, os.Kill)
		for {
			select {
			case <-sigch:
				err = proc.Kill()
				if err != nil {
					log.Fatalf("error killing monitored process: %+v", err)
				}
			}
		}
	}()

	err = proc.Run()
	if err != nil {
		log.Fatalf("error monitoring process: %+v", err)
	}

	err = w.Flush()
	if err != nil {
		log.Fatalf("error flushing log file: %+v", err)
	}

	err = f.Close()
	if err != nil {
		log.Fatalf("error closing log file: %+v", err)
	}
}

func runPID(out string, pid int) {
	f, err := os.Create(out)
	if err != nil {
		log.Fatalf("could not create output log file: %+v", err)
	}
	defer f.Close()

	w := bufio.NewWriter(f)

	proc, err := pmon.Monitor(pid)
	if err != nil {
		log.Fatalf("could not create monitor for process PID=%d: %+v", pid, err)
	}
	proc.W = w
	proc.Freq = *freq

	go func() {
		sigch := make(chan os.Signal, 1)
		signal.Notify(sigch, os.Interrupt, os.Kill)
		for {
			select {
			case <-sigch:
				err = proc.Kill()
				if err != nil {
					log.Fatalf("error killing monitored process: %+v", err)
				}
			}
		}
	}()

	err = proc.Run()
	if err != nil {
		log.Fatalf("error monitoring process: %+v", err)
	}

	err = w.Flush()
	if err != nil {
		log.Fatalf("error flushing log file: %+v", err)
	}

	err = f.Close()
	if err != nil {
		log.Fatalf("error closing log file: %+v", err)
	}
}
