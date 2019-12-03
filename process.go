// Copyright 2015 The pmon Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pmon

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	sys "golang.org/x/sys/unix"
)

type Process struct {
	Cmd  *exec.Cmd
	W    io.Writer
	Freq time.Duration

	quit chan struct{}

	fc chan func() error
	ec chan error
}

func New(cmd string, args ...string) *Process {
	c := exec.Command(cmd, args...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	c.SysProcAttr = &sys.SysProcAttr{
		Ptrace:  true,
		Setpgid: true,
	}

	proc := &Process{
		Cmd:  c,
		Freq: 1 * time.Second,
		W:    ioutil.Discard,
		quit: make(chan struct{}),
		fc:   make(chan func() error),
		ec:   make(chan error),
	}

	go ptraceRun(proc.fc, proc.ec)
	return proc
}

func (p *Process) Run() error {
	err := p.ptraceRun(p.Cmd.Start)
	if err != nil {
		return fmt.Errorf("could not start process: %w", err)
	}

	start := time.Now()

	pid := p.Cmd.Process.Pid
	collector, err := newCollector(pid)
	if err != nil {
		return fmt.Errorf("could not create collector: %w", err)
	}
	defer collector.Close()

	_, err = fmt.Fprintf(p.W,
		"# pmon: %s\n# freq: %v\n# format: %#v\n# start: %v\n",
		strings.Join(p.Cmd.Args, " "),
		p.Freq,
		Infos{},
		start,
	)
	if err != nil {
		return fmt.Errorf("error writing log-file header: %w", err)
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

	err = p.wait(pid, 0)
	if err != nil {
		return fmt.Errorf("waiting for target execve failed: %w", err)
	}

	err = p.ptraceDetach(pid)
	if err != nil {
		return fmt.Errorf("could not ptrace-detach pid=%d: %w", pid, err)
	}

	go p.monitor(collector)

	log.Printf(
		"monitoring... (pid=%d, freq=%v)\n",
		p.Cmd.Process.Pid,
		p.Freq,
	)
	err = p.Cmd.Wait()
	p.quit <- struct{}{}
	if err != nil {
		return fmt.Errorf("could not wait for pid=%d: %w", pid, err)
	}

	return nil
}

func (p *Process) Kill() error {
	pgid, err := sys.Getpgid(p.Cmd.Process.Pid)
	if err != nil {
		return fmt.Errorf("could not get process group of pid=%d: %w", p.Cmd.Process.Pid, err)
	}
	err = sys.Kill(-pgid, sys.SIGKILL) // note the minus sign
	if err != nil {
		return fmt.Errorf("could not kill process group %d: %w", pgid, err)
	}

	return nil
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
		log.Printf("error collecting: %+v", err)
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
