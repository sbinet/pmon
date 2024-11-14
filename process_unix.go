// Copyright 2015 The pmon Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pmon

import (
	"os"
)

const (
	// ClockTicks is the number of clock ticks per second.
	// ClockTicks is a constant on Linux and Darwin.
	ClockTicks = uint64(100) // uint64(C.sysconf(C._SC_CLK_TCK))

	clockTicksToNanosecond = (1000000000 / ClockTicks)
)

var (
	// PageSize is the underlying system's memory page size.
	PageSize = int64(os.Getpagesize())
)
