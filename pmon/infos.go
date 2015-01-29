// Copyright 2015 The pmon Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pmon

import (
	"time"
)

type Infos struct {
	CPU     time.Duration `json:"cpu"`      // user+system time (ms)
	UTime   time.Duration `json:"usr"`      // user time (ms)
	STime   time.Duration `json:"sys"`      // system time (ms)
	VMem    int64         `json:"vmem"`     // virtual memory (kB)
	RSS     int64         `json:"rss"`      // resident set size (kB)
	Threads int64         `json:"nthreads"` // number of threads
	Rchar   int64         `json:"rchar"`    // number of bytes read from storage (kB)
	Wchar   int64         `json:"wchar"`    // number of bytes written to storage (kB)
	Rdisk   int64         `json:"rdisk"`    // number of bytes read from physical storage (kB)
	Wdisk   int64         `json:"wdisk"`    // number of bytes written to physical storage (kB)
}
