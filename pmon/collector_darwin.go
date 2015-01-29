// Copyright 2015 The pmon Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pmon

/*
//#include <stdlib.h>
//#include <sys/sysctl.h>
//#include <sys/mount.h>
#include <mach/mach_init.h>
#include <mach/mach_host.h>
#include <mach/host_info.h>
#include <libproc.h>
#include <mach/processor_info.h>
#include <mach/vm_map.h>
*/
import "C"

import (
	"syscall"
	"time"
	"unsafe"
)

type collector struct {
	pid int
}

func newCollector(pid int) (*collector, error) {
	return &collector{pid: pid}, nil
}

func (c *collector) Close() error {
	var err error
	return err
}

func (c *collector) collect() (Infos, error) {

	info := C.struct_proc_taskallinfo{}

	err := task_info(c.pid, &info)
	if err != nil {
		return Infos{}, err
	}

	var (
		usr = time.Duration(info.ptinfo.pti_total_user) / time.Millisecond
		sys = time.Duration(info.ptinfo.pti_total_system) / time.Millisecond
	)
	infos := Infos{
		CPU:     usr + sys,
		UTime:   usr,
		STime:   sys,
		VMem:    int64(info.ptinfo.pti_virtual_size / 1024),  // in kB
		RSS:     int64(info.ptinfo.pti_resident_size / 1024), // in kB
		Threads: int64(info.ptinfo.pti_threadnum),
		Rchar:   -1,
		Wchar:   -1,
		Rdisk:   -1,
		Wdisk:   -1,
	}

	return infos, err
}

func task_info(pid int, info *C.struct_proc_taskallinfo) error {
	size := C.int(unsafe.Sizeof(*info))
	ptr := unsafe.Pointer(info)

	n := C.proc_pidinfo(C.int(pid), C.PROC_PIDTASKALLINFO, 0, ptr, size)
	if n != size {
		return syscall.ENOMEM
	}

	return nil
}
