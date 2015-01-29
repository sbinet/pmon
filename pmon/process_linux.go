// Copyright 2015 The pmon Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pmon

import (
	"syscall"
)

func wait(pid, options int) (int, *syscall.WaitStatus, error) {
	var status syscall.WaitStatus
	wpid, err := syscall.Wait4(pid, &status, syscall.WALL|options, nil)
	return wpid, &status, err
}