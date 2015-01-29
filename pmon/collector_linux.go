// Copyright 2015 The pmon Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pmon

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type collector struct {
	stat *os.File
	io   *os.File
}

func newCollector(pid int) (*collector, error) {
	dir := "/proc/" + strconv.Itoa(pid)
	stat, err := os.Open(dir + "/stat")
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not open /proc/%d/stat: %v\n", pid, err)
		return nil, err
	}

	io, err := os.Open(dir + "/io")
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not open /proc/%d/io: %v\n", pid, err)
		return nil, err
	}

	return &collector{stat: stat, io: io}, nil
}

func (c *collector) Close() error {
	err1 := c.stat.Close()
	err2 := c.io.Close()
	if err1 != nil {
		return err1
	}
	if err2 != nil {
		return err2
	}
	return nil
}

const (
	statfmt = "%d %s %c %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d"
)

func (c *collector) collect() (Infos, error) {
	// see: http://man7.org/linux/man-pages/man5/proc.5.html
	stat := struct {
		pid       int    // process ID
		comm      string // filename of the executable in parentheses
		state     byte   // process state
		ppid      int    // pid of the parent process
		pgrp      int    // process group ID of the process
		session   int    // session ID of the process
		tty       int    // controlling terminal of the process
		tpgid     int    // ID of foreground process group
		flags     uint32 // kernel flags word of the process
		minflt    uint64 // number of minor faults the process has made which have not required loading a memory page from disk
		cminflt   uint64 // number of minor faults the process's waited-for children have made
		majflt    uint64 // number of major faults the process has made which have required loading a memory page from disk
		cmajflt   uint64 // number of major faults the process's waited-for children have made
		utime     uint64 // user time in clock ticks
		stime     uint64 // system time in clock ticks
		cutime    int64  // children user time in clock ticks
		cstime    int64  // children system time in clock ticks
		priority  int64  // priority
		nice      int64  // the nice value
		nthreads  int64  // number of threads in this process
		itrealval int64  // time in jiffies before next SIGALRM is sent to the process dure to an interval timer
		starttime int64  // time the process started after system boot in clock ticks
		vsize     uint64 // virtual memory size in bytes
		rss       int64  // resident set size: number of pages the process has in real memory
	}{}

	_, err := c.stat.Seek(0, 0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not rewind %s: %v\n", c.stat.Name, err)
		return Infos{}, err
	}

	_, err = fmt.Fscanf(
		c.stat, statfmt,
		&stat.pid, &stat.comm, &stat.state,
		&stat.ppid, &stat.pgrp, &stat.session,
		&stat.tty, &stat.tpgid, &stat.flags,
		&stat.minflt, &stat.cminflt, &stat.majflt, &stat.cmajflt,
		&stat.utime, &stat.stime,
		&stat.cutime, &stat.cstime,
		&stat.priority,
		&stat.nice,
		&stat.nthreads,
		&stat.itrealval, &stat.starttime,
		&stat.vsize, &stat.rss,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error collecting CPU/Mem data: %v\n", err)
		return Infos{}, err
	}

	_, err = c.io.Seek(0, 0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not rewind %s: %v\n", c.io.Name, err)
		return Infos{}, err
	}

	var (
		rchar int64
		wchar int64
		syscr int64
		syscw int64
		rdisk int64
		wdisk int64
	)
	_, err = fmt.Fscanf(
		c.io,
		"rchar: %d\nwchar: %d\nsyscr: %d\nsyscw: %d\nread_bytes: %d\nwrite_bytes: %d\n",
		&rchar, &wchar,
		&syscr, &syscw,
		&rdisk, &wdisk,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error collecting I/O data: %v\n", err)
		return Infos{}, err
	}

	infos := Infos{
		CPU:     time.Duration((stat.utime + stat.stime) * clockTicksToNanosecond),
		UTime:   time.Duration(stat.utime * clockTicksToNanosecond),
		STime:   time.Duration(stat.stime * clockTicksToNanosecond),
		VMem:    int64(stat.vsize) / 1024,   // in kB
		RSS:     stat.rss * PageSize / 1024, // in kB
		Threads: stat.nthreads,
		Rchar:   rchar / 1024, // in kB
		Wchar:   wchar / 1024, // in kB
		Rdisk:   rdisk / 1024, // in kB
		Wdisk:   wdisk / 1024, // in kB
	}
	return infos, err
}

/*
syscr: 6
syscw: 0
read_bytes: 0
write_bytes: 0

*/
