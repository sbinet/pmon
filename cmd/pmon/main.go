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
	//mon  = flag.String("mon", "cpu,vmem,rss", "comma-separated list of resources to monitor")
	out = flag.String("o", "pmon.data", "path to file to store resources usage log")

	usage = `pmon monitors process resources usage.

Usage:

 $ pmon [options] command [command-arg1 [command-arg2 [...]]]

Example:

 $ pmon my-command arg0 arg1
 $ pmon -- my-command arg0 arg1

Options:
`
)

func main() {

	log.SetFlags(0)
	log.SetPrefix("pmon: ")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, usage)
		flag.PrintDefaults()
	}

	flag.Parse()

	if flag.NArg() <= 0 {
		log.Printf("expect a command (and its arguments) as argument")
		flag.Usage()
		os.Exit(1)
	}

	err := os.MkdirAll(filepath.Dir(*out), 0755)
	if err != nil {
		log.Fatalf("could not create output directory: %+v", err)
	}

	cmd := flag.Arg(0)
	args := flag.Args()[1:]

	run(*out, cmd, args)
}

func run(out, cmd string, args []string) {
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
		sigch := make(chan os.Signal)
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
