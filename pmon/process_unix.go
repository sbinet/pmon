// Copyright 2015 The pmon Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pmon

import (
	"os"
)

var (
	ClockTicks = uint64(100) // uint64(C.sysconf(C._SC_CLK_TCK))
	PageSize   = int64(os.Getpagesize())

	clockTicksToNanosecond = (1000000000 / ClockTicks)
)
