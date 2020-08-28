// Copyright 2015 The pmon Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pmon

import (
	sys "golang.org/x/sys/unix"
)

func (p *Process) wait(pid, options int) error {
	p.fc <- func() error {
		var status sys.WaitStatus
		_, err := sys.Wait4(pid, &status, options, nil)
		return err
	}
	return <-p.ec
}

func (p *Process) cmdline(pid int) string {
	// TODO(sbinet)
	return "<N/A>"
}
