// Copyright 2015 The pmon Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pmon

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
)

// Infos holds monitoring informations gathered during monitoring.
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

// Meta holds metadata about a pmon run.
type Meta struct {
	Cmd     string
	Freq    time.Duration
	Start   time.Time
	Elapsed time.Duration
	Stop    time.Time

	Infos []Infos
}

// Parse parses a pmon run log file.
func Parse(r io.Reader) (Meta, error) {
	const (
		layout = time.RFC3339Nano
		ms2sec = 1 / 1000.0
	)

	var (
		meta Meta
		err  error
	)

	sc := bufio.NewScanner(r)
	for sc.Scan() {
		txt := strings.TrimSpace(sc.Text())
		if txt == "" {
			continue
		}

		switch txt[0] {
		case '#':
			switch {
			case strings.HasPrefix(txt, "# pmon: "):
				meta.Cmd = string(txt[len("# pmon: "):])

			case strings.HasPrefix(txt, "# freq: "):
				v, err := time.ParseDuration(txt[len("# freq: "):])
				if err != nil {
					return meta, fmt.Errorf("could not parse frequency %q: %w", txt, err)
				}
				meta.Freq = v

			case strings.HasPrefix(txt, "# start: "):
				v, err := time.Parse(layout, txt[len("# start: "):])
				if err != nil {
					return meta, fmt.Errorf("could not parse start time %q: %w", txt, err)
				}
				meta.Start = v

			case strings.HasPrefix(txt, "# elapsed: "):
				v, err := time.ParseDuration(txt[len("# elapsed: "):])
				if err != nil {
					return meta, fmt.Errorf("could not parse elapsed time %q: %w", txt, err)
				}
				meta.Elapsed = v

			case strings.HasPrefix(txt, "# stop: "):
				v, err := time.Parse(layout, txt[len("# stop: "):])
				if err != nil {
					return meta, fmt.Errorf("could not parse stop time %q: %w", txt, err)
				}
				meta.Stop = v

			}
		default:
			var (
				v   Infos
				cpu float64
				usr float64
				sys float64
			)
			_, e := fmt.Sscanf(txt, "%f %f %f %d %d %d %d %d %d %d",
				&cpu, &usr, &sys, &v.VMem, &v.RSS,
				&v.Threads,
				&v.Rchar, &v.Wchar,
				&v.Rdisk, &v.Wdisk,
			)
			if e != nil {
				err = errors.Join(err, fmt.Errorf("could not scan pmon-info %q: %w", txt, e))
				continue
			}
			v.CPU = time.Duration(int(cpu * ms2sec))
			v.UTime = time.Duration(int(usr * ms2sec))
			v.STime = time.Duration(int(sys * ms2sec))
			meta.Infos = append(meta.Infos, v)
		}
	}

	if e := sc.Err(); e != nil {
		err = errors.Join(err, fmt.Errorf("could not scan input reader: %w", e))
	}

	return meta, err
}
