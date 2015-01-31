// Copyright 2015 The pmon Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pmon

import (
	"syscall"
)

func (p *Process) wait(pid, options int) error {
	p.fc <- func() error {
		var err1 error
		var status syscall.WaitStatus
		_, err1 = syscall.Wait4(pid, &status, options, nil)
		return err1
	}
	return <-p.ec
}
