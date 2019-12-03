// Copyright 2015 The pmon Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pmon

import (
	"runtime"
	"syscall"
)

// ptraceRun runs all the closures from fc on a dedicated OS thread. Errors
// are returned on ec. Both channels must be unbuffered, to ensure that the
// resultant error is sent back to the same goroutine that sent the closure.
func ptraceRun(fc chan func() error, ec chan error) {
	if cap(fc) != 0 || cap(ec) != 0 {
		panic("ptraceRun was given buffered channels")
	}
	runtime.LockOSThread()
	for f := range fc {
		ec <- f()
	}
}

func (p *Process) ptraceDetach(pid int) error {
	p.fc <- func() error {
		return syscall.PtraceDetach(pid)
	}
	return <-p.ec
}

func (p *Process) ptraceRun(f func() error) error {
	p.fc <- f
	return <-p.ec
}
