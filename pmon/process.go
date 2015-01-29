// Copyright 2015 The pmon Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// package pmon monitors the resources of a process.
package pmon

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

type Process struct {
	Cmd  *exec.Cmd
	W    io.Writer
	Freq time.Duration

	quit chan struct{}
}

func New(cmd string, args ...string) *Process {
	c := exec.Command(cmd, args...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	c.SysProcAttr = &syscall.SysProcAttr{
		Ptrace:  true,
		Setpgid: true,
	}

	return &Process{
		Cmd:  c,
		Freq: 1 * time.Second,
		W:    ioutil.Discard,
		quit: make(chan struct{}),
	}
}

func (p *Process) Run() error {
	err := p.Cmd.Start()
	if err != nil {
		fmt.Fprintf(os.Stderr, "start-error: %v\n", err)
		return err
	}

	start := time.Now()

	pid := p.Cmd.Process.Pid
	collector, err := newCollector(pid)
	if err != nil {
		return fmt.Errorf("error creating collector: %v\n", err)
	}
	defer collector.Close()

	_, err = fmt.Fprintf(p.W,
		"# pmon: %s %s\n# freq: %v\n# format: %#v\n# start: %v\n",
		p.Cmd.Path, strings.Join(p.Cmd.Args, " "),
		p.Freq,
		Infos{},
		start,
	)
	if err != nil {
		return fmt.Errorf("error writing log-file header: %v\n", err)
	}

	defer func() {
		stop := time.Now()
		delta := time.Since(start)
		_, _ = fmt.Fprintf(p.W,
			"# elapsed: %v\n# stop: %v\n",
			delta,
			stop,
		)
	}()

	_, _, err = wait(pid, 0)
	if err != nil {
		return fmt.Errorf("waiting for target execve failed: %s", err)
	}

	err = syscall.PtraceDetach(pid)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ptrace-error: %v\n\nretry...\n", err)
		time.Sleep(1 * time.Second)
		err = syscall.PtraceDetach(pid)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ptrace-error: %v\n", err)
			return err
		}
	}

	go p.monitor(collector)

	fmt.Fprintf(os.Stderr,
		"pmon: monitoring... (pid=%d, freq=%v)\n",
		p.Cmd.Process.Pid,
		p.Freq,
	)
	err = p.Cmd.Wait()
	p.quit <- struct{}{}
	return err
}

func (p *Process) Kill() error {
	pgid, err := syscall.Getpgid(p.Cmd.Process.Pid)
	if err != nil {
		return err
	}
	err = syscall.Kill(-pgid, syscall.SIGKILL) // note the minus sign
	return err
}

func (p *Process) monitor(c *collector) {
	p.collect(c)
	tick := time.Tick(p.Freq)
	for {
		select {
		case <-tick:
			p.collect(c)
		case <-p.quit:
			return
		}
	}
}

func (p *Process) collect(c *collector) {

	if p.Cmd.ProcessState != nil {
		// process already stopped. nothing to collect.
		return
	}

	infos, err := c.collect()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error collecting: %v\n", err)
		return
	}

	fmt.Fprintf(
		p.W, "%f %f %f %d %d %d %d %d %d %d\n",
		milliseconds(infos.CPU), milliseconds(infos.UTime), milliseconds(infos.STime),
		infos.VMem, infos.RSS,
		infos.Threads,
		infos.Rchar, infos.Wchar,
		infos.Rdisk, infos.Wdisk,
	)
}

func milliseconds(t time.Duration) float64 {
	return t.Seconds() * 1e3
}
