// SiFive FU540 initialization
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package fu530 provides support to Go bare metal unikernels written using the
// TamaGo framework.
//
// The package implements initialization and drivers for specific SiFive FU540
// System-on-Chip (SoC) peripherals, adopting, where indicated, the following
// reference specifications:
//   * FU540C00RM - SiFive FU540-C000 Manual - v1p4 2021/03/25
//
// This package is only meant to be used with `GOOS=tamago GOARCH=riscv64` as
// supported by the TamaGo framework for bare metal Go on RISC-V SoCs, see
// https://github.com/usbarmory/tamago.
package fu540

import (
	_ "unsafe"

	"github.com/usbarmory/tamago/soc/fu540/uart"
)

// Peripheral registers
const (
	CLINT_BASE = 0x2000000
	MTIME      = 0xbff8

	UART1_BASE = 0x10010000
	UART2_BASE = 0x10011000
)

// Peripheral instances
var (
	UART1 = &uart.UART{
		Index: 1,
		Base:  UART1_BASE,
	}

	UART2 = &uart.UART{
		Index: 2,
		Base:  UART2_BASE,
	}
)

//go:linkname ramStackOffset runtime.ramStackOffset
var ramStackOffset uint64 = 0x100

// Init takes care of the lower level SoC initialization triggered early in
// runtime setup.
func Init() {
	return
}

// Model returns the SoC model name.
func Model() string {
	return "FU540"
}