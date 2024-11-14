// Copyright 2020 The pmon Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pmon

import (
	"fmt"
	"io"
	"os"
	"time"
)

// Monitor monitors the resources usage of the process with the provided id.
func Monitor(pid int) (*Process, error) {
	p, err := os.FindProcess(pid)
	if err != nil {
		return nil, fmt.Errorf("could not find process with pid=%d: %w", pid, err)
	}

	proc := &Process{
		Freq:  1 * time.Second,
		W:     io.Discard,
		quit:  make(chan struct{}),
		start: func() error { return nil },

		proc: p,
	}

	proc.stop = func() error {
		close(proc.quit)
		return nil
	}

	return proc, nil
}
