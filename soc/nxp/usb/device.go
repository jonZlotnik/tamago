// USB device mode support
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package usb

import (
	"log"
	"sync"
	"time"

	"github.com/usbarmory/tamago/internal/reg"
)

// DeviceMode sets the USB controller in device mode.
func (hw *USB) DeviceMode() {
	hw.Lock()
	defer hw.Unlock()

	reg.Set(hw.cmd, USBCMD_RST)
	reg.Wait(hw.cmd, USBCMD_RST, 1, 0)

	// p3872, 56.6.33 USB Device Mode (USB_nUSBMODE), IMX6ULLRM)
	m := reg.Read(hw.mode)

	// set device only controller
	m = (m & ^uint32(0b11<<USBMODE_CM)) | (USBMODE_CM_DEVICE << USBMODE_CM)
	// disable setup lockout
	m |= (1 << USBMODE_SLOM)
	// disable stream mode
	m &^= (1 << USBMODE_SDIS)

	reg.Write(hw.mode, m)
	reg.Wait(hw.mode, USBMODE_CM, 0b11, USBMODE_CM_DEVICE)

	// initialize endpoint queue head list
	hw.initQH()
	// set control endpoint
	hw.set(0, IN, 64, true, 0)
	hw.set(0, OUT, 64, true, 0)

	// set OTG termination
	reg.Set(hw.otg, OTGSC_OT)

	// clear all pending interrupts
	reg.Write(hw.sts, 0xffffffff)

	// run
	reg.Set(hw.cmd, USBCMD_RS)
}

// Start waits and handles configured USB endpoints in device mode, it should
// never return. Note that isochronous endpoints are not supported.
func (hw *USB) Start(dev *Device) {
	var conf uint8
	var wg sync.WaitGroup

	for {
		// check for bus reset
		if reg.Get(hw.sts, USBSTS_URI, 1) == 1 {
			// set inactive configuration
			conf = 0
			dev.ConfigurationValue = 0

			// perform controller reset procedure
			hw.Reset()
			log.Println("RESET DONE")
		}

		// wait for a setup packet
		if !reg.WaitFor(10*time.Millisecond, hw.setup, 0, 1, 1) {
			log.Println("Waiting for setup...")
			continue
		}

		// handle setup packet
		s := hw.getSetup()
		log.Println("RETURNED from hw.getSetup")
		if err := hw.handleSetup(dev, s); err != nil {
			log.Printf("usb: setup error, %v", err)
		}
		log.Println("RETURNED from hw.handleSetup")

		// check if configuration reload is required
		if dev.ConfigurationValue == conf {
			log.Println("Config reload required")
			continue
		} else {
			// Host has chosen a configuration from dev.Configurations
			// Save choice to start endpoints from this config
			conf = dev.ConfigurationValue
		}

		// stop configuration endpoints
		if hw.done != nil {
			log.Println("CLOSING hw.done")
			close(hw.done)
			wg.Wait()
		}
		// start configuration endpoints
		log.Println("STARTING ENDPOINTS")
		hw.startEndpoints(&wg, dev, conf)
		log.Println("RETURNED from startEndpoints")
	}
}
