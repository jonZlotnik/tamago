// USB armory Mk II support for tamago/arm
// https://github.com/f-secure-foundry/tamago
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package usbarmory

import (
	"github.com/f-secure-foundry/tamago/soc/imx6"
	"github.com/f-secure-foundry/tamago/soc/imx6/usdhc"
)

// SD instance
var SD = usdhc.USDHC1

// MMC instance
var MMC = usdhc.USDHC2

// SD/MMC configuration constants.
//
// On the USB armory Mk II the following uSDHC interfaces are connected:
//   * uSDHC1: external uSD  slot (SD1)
//   * uSDHC2: internal eMMC card (SD2/NAND)
//
// On the USB armory Mk II β revision the maximum achievable theoretical speed
// modes are:
//   * uSD:  High Speed (HS)      25MB/s, 50MHz, 3.3V, 4-bit data bus
//   * eMMC: High Speed (HS) DDR 104MB/s, 52MHz, 3.3V, 8-bit data bus
//
// On the USB armory Mk II γ revision the maximum achievable theoretical speed
// modes are:
//   * uSD:  SDR104  75MB/s, 150MHz, 1.8V, 4-bit data bus
//   * eMMC: HS200  150MB/s, 150MHz, 1.8V, 8-bit data bus
const (
	IOMUXC_SW_MUX_CTL_PAD_CSI_DATA04 = 0x020e01f4
	IOMUXC_SW_PAD_CTL_PAD_CSI_DATA04 = 0x020e0480
	IOMUXC_USDHC1_WP_SELECT_INPUT    = 0x020e066c

	USDHC1_WP_MODE   = 8
	DAISY_CSI_DATA04 = 0b10

	IOMUXC_SW_MUX_CTL_PAD_CSI_PIXCLK = 0x020e01d8
	IOMUXC_SW_PAD_CTL_PAD_CSI_PIXCLK = 0x020e0464
	IOMUXC_USDHC2_WP_SELECT_INPUT    = 0x020e069c

	USDHC2_WP_MODE   = 1
	DAISY_CSI_PIXCLK = 0b10

	SD_BUS_WIDTH  = 4
	MMC_BUS_WIDTH = 8

	PF1510_LDO3_VOLT = 0x52
	LDO3_VOLT_1V8    = 0x10
	LDO3_VOLT_3V3    = 0x1f
)

func init() {
	// There are no write-protect lines on uSD or eMMC cards, therefore the
	// respective SoC pads must be selected on pulled down unconnected pads
	// to ensure the driver never sees write protection enabled.
	ctl := uint32((1 << imx6.SW_PAD_CTL_PUE) | (1 << imx6.SW_PAD_CTL_PKE))

	// SD write protect (USDHC1_WP)
	wpSD, err := imx6.NewPad(IOMUXC_SW_MUX_CTL_PAD_CSI_DATA04,
		IOMUXC_SW_PAD_CTL_PAD_CSI_DATA04,
		IOMUXC_USDHC1_WP_SELECT_INPUT)

	if err != nil {
		panic(err)
	}

	// MMC write protect (USDHC2_WP)
	wpMMC, err := imx6.NewPad(IOMUXC_SW_MUX_CTL_PAD_CSI_PIXCLK,
		IOMUXC_SW_PAD_CTL_PAD_CSI_PIXCLK,
		IOMUXC_USDHC2_WP_SELECT_INPUT)

	if err != nil {
		panic(err)
	}

	if !imx6.Native {
		return
	}

	wpSD.Mode(USDHC1_WP_MODE)
	wpSD.Select(DAISY_CSI_DATA04)
	wpSD.Ctl(ctl)

	wpMMC.Mode(USDHC2_WP_MODE)
	wpMMC.Select(DAISY_CSI_PIXCLK)
	wpMMC.Ctl(ctl)

	SD.Init(SD_BUS_WIDTH)
	MMC.Init(MMC_BUS_WIDTH)

	switch Model() {
	case "UA-MKII-β":
		// β revisions do not support SDR104 (SD) or HS200 (MMC)
		return
	case "UA-MKII-γ":
		SD.LowVoltage = lowVoltageSD
		MMC.LowVoltage = lowVoltageMMC
	}
}

func lowVoltageSD(enable bool) bool {
	a := make([]byte, 1)

	if enable {
		a[0] = LDO3_VOLT_1V8
	} else {
		a[0] = LDO3_VOLT_3V3
	}

	err := imx6.I2C1.Write(a, PF1510_ADDR, PF1510_LDO3_VOLT, 1)

	return err == nil
}

func lowVoltageMMC(enable bool) bool {
	return true
}
