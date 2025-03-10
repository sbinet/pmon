// Copyright 2015 The pmon Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pmon

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	sys "golang.org/x/sys/unix"
)

// Process holds informations about a process created by New or Monitor.
type Process struct {
	W    io.Writer
	Freq time.Duration

	quit chan struct{}

	fc chan func() error
	ec chan error

	Msg  *log.Logger
	Cmd  *exec.Cmd
	proc *os.Process

	start func() error
	stop  func() error
}

// New creates a new process named cmd and with the provided arguments.
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
		Msg:  log.Default(),
		Cmd:  c,
		Freq: 1 * time.Second,
		W:    io.Discard,
		quit: make(chan struct{}),
		fc:   make(chan func() error),
		ec:   make(chan error),

		start: c.Start,
		stop: func() error {
			pgid, err := sys.Getpgid(c.Process.Pid)
			if err != nil {
				return fmt.Errorf("could not get process group of pid=%d: %w", c.Process.Pid, err)
			}
			err = sys.Kill(-pgid, sys.SIGKILL) // note the minus sign
			if err != nil {
				return fmt.Errorf("could not kill process group %d: %w", pgid, err)
			}

			return nil
		},
	}

	go ptraceRun(proc.fc, proc.ec)
	return proc
}

// Run starts the monitoring of the current process.
func (p *Process) Run() error {
	switch {
	case p.Cmd != nil:
		return p.runCmd()
	default:
		return p.runPID()
	}
}

func (p *Process) runCmd() error {
	defer close(p.quit)
	defer func() {
		if w, ok := p.W.(interface{ Flush() error }); ok {
			_ = w.Flush()
		}
	}()

	err := p.ptraceRun(p.Cmd.Start)
	if err != nil {
		return fmt.Errorf("could not start process: %w", err)
	}

	start := time.Now()

	pid := p.Cmd.Process.Pid
	collector, err := newCollector(p.Msg, pid)
	if err != nil {
		return fmt.Errorf("could not create collector: %w", err)
	}
	defer collector.Close()

	_, err = fmt.Fprintf(p.W,
		"# pmon: %s\n# freq: %v\n# format: %#v\n# start: %v\n",
		strings.Join(p.Cmd.Args, " "),
		p.Freq,
		Infos{},
		start.Format(time.RFC3339Nano),
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
			stop.Format(time.RFC3339Nano),
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

	p.Msg.Printf(
		"monitoring... (pid=%d, freq=%v)\n",
		p.Cmd.Process.Pid,
		p.Freq,
	)
	err = p.Cmd.Wait()
	if err != nil {
		return fmt.Errorf("could not wait for pid=%d: %w", pid, err)
	}

	return nil
}

func (p *Process) runPID() error {
	start := time.Now()

	pid := p.proc.Pid
	collector, err := newCollector(p.Msg, pid)
	if err != nil {
		return fmt.Errorf("could not create collector: %w", err)
	}
	defer collector.Close()

	_, err = fmt.Fprintf(p.W,
		"# pmon: %s\n# freq: %v\n# format: %#v\n# start: %v\n",
		p.cmdline(pid),
		p.Freq,
		Infos{},
		start.Format(time.RFC3339Nano),
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
			stop.Format(time.RFC3339Nano),
		)
	}()

	go p.monitor(collector)

	p.Msg.Printf(
		"monitoring... (pid=%d, freq=%v)\n",
		p.proc.Pid,
		p.Freq,
	)
	<-p.quit
	if err != nil {
		return fmt.Errorf("could not wait for pid=%d: %w", pid, err)
	}

	return nil
}

// Kill causes the monitored process to exit immediately.
func (p *Process) Kill() error {
	return p.stop()
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

	if p.Cmd != nil && p.Cmd.ProcessState != nil {
		// process already stopped. nothing to collect.
		return
	}

	infos, err := c.collect()
	if err != nil {
		p.Msg.Printf("error collecting: %+v", err)
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
