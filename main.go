// Copyright 2015 The pmon Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path"
	"time"

	"github.com/sbinet/pmon/pmon"
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

func printf(format string, args ...interface{}) (int, error) {
	return fmt.Fprintf(os.Stderr, "pmon: "+format, args...)
}

func fatalf(format string, args ...interface{}) {
	printf(format, args...)
	os.Exit(1)
}

func main() {

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, usage)
		flag.PrintDefaults()
	}

	flag.Parse()

	if flag.NArg() <= 0 {
		printf("expect a command (and its arguments) as argument\n")
		flag.Usage()
		os.Exit(1)
	}

	err := os.MkdirAll(path.Dir(*out), 0755)
	if err != nil {
		fatalf("could not create output directory: %v\n", err)
	}

	cmd := flag.Arg(0)
	args := flag.Args()[1:]

	f, err := os.Create(*out)
	if err != nil {
		fatalf("could not create output log file: %v\n", err)
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
					fatalf("error killing monitored process: %v\n", err)
				}
			}
		}
	}()

	err = proc.Run()
	if err != nil {
		fatalf("error monitoring process: %v\n", err)
	}

	err = w.Flush()
	if err != nil {
		fatalf("error flushing log file: %v\n", err)
	}

	err = f.Close()
	if err != nil {
		fatalf("error closing log file: %v\n", err)
	}

	os.Exit(0)
}
