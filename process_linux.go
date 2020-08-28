// Copyright 2015 The pmon Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pmon

import (
	"fmt"
	"io/ioutil"

	sys "golang.org/x/sys/unix"
)

func (p *Process) wait(pid, options int) error {
	p.fc <- func() error {
		var status sys.WaitStatus
		_, err := sys.Wait4(pid, &status, sys.WALL|options, nil)
		return err
	}
	return <-p.ec
}

func (p *Process) cmdline(pid int) string {
	cmd, err := ioutil.ReadFile(fmt.Sprintf("/proc/%d/cmdline", pid))
	if err != nil {
		return "<N/A>"
	}
	return string(cmd)
}
