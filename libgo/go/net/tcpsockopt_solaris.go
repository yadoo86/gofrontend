// Copyright 2009 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package net

import (
	"os"
	"syscall"
	"time"
)

// Set keep alive period.
func setKeepAlivePeriod(fd *netFD, d time.Duration) error {
	if err := fd.incref(); err != nil {
		return err
	}
	defer fd.decref()

	// The kernel expects milliseconds so round to next highest millisecond.
	d += (time.Millisecond - time.Nanosecond)
	msecs := int(d.Nanoseconds() / time.Millisecond)

	return os.NewSyscallError("setsockopt", syscall.SetsockoptInt(fd.sysfd, syscall.IPPROTO_TCP, syscall.TCP_KEEPALIVE_THRESHOLD, msecs))
}
