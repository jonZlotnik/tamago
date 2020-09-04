// BCM2835 mini-UART
// https://github.com/f-secure-foundry/tamago
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//
// This mini-UART is specifically intended for use as a
// console.  See BCM2835-ARM-Peripherals.pdf that is
// widely available.
//

package bcm2835

import (
	// uses go:linkname

	_ "unsafe"

	"github.com/f-secure-foundry/tamago/arm"
	"github.com/f-secure-foundry/tamago/internal/reg"
)

//go:linkname printk runtime.printk
func printk(c byte) {
	MiniUART.Tx(c)
}

const (
	AUX_ENABLES     = 0x215004
	AUX_MU_IO_REG   = 0x215040
	AUX_MU_IER_REG  = 0x215044
	AUX_MU_IIR_REG  = 0x215048
	AUX_MU_LCR_REG  = 0x21504C
	AUX_MU_MCR_REG  = 0x215050
	AUX_MU_LSR_REG  = 0x215054
	AUX_MU_MSR_REG  = 0x215058
	AUX_MU_SCRATCH  = 0x21505C
	AUX_MU_CNTL_REG = 0x215060
	AUX_MU_STAT_REG = 0x215064
	AUX_MU_BAUD_REG = 0x215068

	GPFSEL1   = 0x200004
	GPPUD     = 0x200094
	GPPUDCLK0 = 0x200098
)

type miniUART struct {
}

// MiniUART is a secondary low throughput UART intended to be
// used as a console.
var MiniUART = &miniUART{}

func (hw *miniUART) Init() {
	reg.Write(PeripheralAddress(AUX_ENABLES), 1)
	reg.Write(PeripheralAddress(AUX_MU_IER_REG), 0)
	reg.Write(PeripheralAddress(AUX_MU_CNTL_REG), 0)
	reg.Write(PeripheralAddress(AUX_MU_LCR_REG), 3)
	reg.Write(PeripheralAddress(AUX_MU_MCR_REG), 0)
	reg.Write(PeripheralAddress(AUX_MU_IER_REG), 0)
	reg.Write(PeripheralAddress(AUX_MU_IIR_REG), 0xC6)
	reg.Write(PeripheralAddress(AUX_MU_BAUD_REG), 270)

	// Not using GPIO abstraction here because at the point
	// we initialize mini-UART during initialization, to
	// provide 'console', calling Lock on sync.Mutex fails.
	ra := reg.Read(PeripheralAddress(GPFSEL1))
	ra &= ^(uint32(7) << 12) // gpio14
	ra |= 2 << 12            // alt5
	ra &= ^(uint32(7) << 15) // gpio15
	ra |= 2 << 15            // alt5
	reg.Write(PeripheralAddress(GPFSEL1), ra)

	reg.Write(PeripheralAddress(GPPUD), 0)
	arm.Busyloop(150)
	reg.Write(PeripheralAddress(GPPUDCLK0), (1<<14)|(1<<15))
	arm.Busyloop(150)
	reg.Write(PeripheralAddress(GPPUDCLK0), 0)

	reg.Write(PeripheralAddress(AUX_MU_CNTL_REG), 3)
}

func (hw *miniUART) Tx(c byte) {
	for {
		if (reg.Read(PeripheralAddress(AUX_MU_LSR_REG)) & 0x20) != 0 {
			break
		}
	}

	reg.Write(PeripheralAddress(AUX_MU_IO_REG), uint32(c))
}

// Write data from buffer to serial port.
func (hw *miniUART) Write(buf []byte) {
	for i := 0; i < len(buf); i++ {
		hw.Tx(buf[i])
	}
}

func (hw *miniUART) setupGPIO(num int) {
	gpio, err := NewGPIO(num)
	if err != nil {
		panic(err)
	}

	// MiniUART is exposed as Alt5 on the UART lines
	gpio.SelectFunction(GPIOFunctionAltFunction5)

	// Ensure no pull-up / pull-down has persisted
	gpio.PullUpDown(GPIONoPullupOrPullDown)
}
