// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package reg

import (
	"sync/atomic"
	"unsafe"
)

func Read64(addr uint64) uint64 {
	reg := (*uint64)(unsafe.Pointer(uintptr(addr)))
	return atomic.LoadUint64(reg)
}

func Write64(addr uint64, val uint64) {
	reg := (*uint64)(unsafe.Pointer(uintptr(addr)))
	atomic.StoreUint64(reg, val)
}
