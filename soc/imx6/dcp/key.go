// NXP Data Co-Processor (DCP) driver
// https://github.com/f-secure-foundry/tamago
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package dcp

import (
	"crypto/aes"
	"encoding/binary"
	"errors"

	"github.com/f-secure-foundry/tamago/bits"
	"github.com/f-secure-foundry/tamago/dma"
	"github.com/f-secure-foundry/tamago/internal/reg"
	"github.com/f-secure-foundry/tamago/soc/imx6"
)

// DeriveKeyMemory represents the DMA memory region used for exchanging DCP
// derived keys when the derivation index points to an internal DCP key RAM
// slot.
//
// The default value allocates a DMA region within the i.MX6 On-Chip RAM
// (OCRAM/iRAM) to avoid passing through external RAM.
//
// The DeriveKey() function uses DeriveKeyMemory only when the default DMA
// region is not already set within iRAM.
//
// Applications can override the region with an arbitrary one when the iRAM
// needs to be avoided or is already used as non-default DMA region.
var DeriveKeyMemory = &dma.Region{
	Start: imx6.IRAMStart,
	Size:  imx6.IRAMSize,
}

func init() {
	DeriveKeyMemory.Init()
}

// DeriveKey derives a hardware unique key in a manner equivalent to PKCS#11
// C_DeriveKey with CKM_AES_CBC_ENCRYPT_DATA.
//
// The diversifier is AES-CBC encrypted using the internal OTPMK key (when SNVS
// is enabled).
//
// *WARNING*: when SNVS is not enabled a default non-unique test vector is used
// and therefore key derivation is *unsafe*, see imx6.SNVS().
//
// A negative index argument results in the derived key being computed and
// returned.
//
// An index argument equal or greater than 0 moves the derived key directly to
// the corresponding internal DCP key RAM slot (see SetKey()). This is
// accomplished through an iRAM reserved DMA buffer, to ensure that the key is
// never exposed to external RAM or the Go runtime. In this case no key is
// returned by the function.
func DeriveKey(diversifier []byte, iv []byte, index int) (key []byte, err error) {
	if len(iv) != aes.BlockSize {
		return nil, errors.New("invalid IV size")
	}

	// prepare diversifier for in-place encryption
	key = pad(diversifier, false)

	region := dma.Default()

	if index >= 0 {
		// force use of iRAM if not already set as default DMA region
		if !(region.Start > imx6.IRAMStart && region.Start < imx6.IRAMStart+imx6.IRAMSize) {
			region = DeriveKeyMemory
		}
	}

	pkt := &WorkPacket{}
	pkt.SetCipherDefaults()

	// Use device-specific hardware key for encryption.
	pkt.Control0 |= 1 << DCP_CTRL0_CIPHER_ENCRYPT
	pkt.Control0 |= 1 << DCP_CTRL0_OTP_KEY
	pkt.Control1 |= KEY_SELECT_UNIQUE_KEY << DCP_CTRL1_KEY_SELECT

	pkt.BufferSize = uint32(len(key))

	pkt.SourceBufferAddress = region.Alloc(key, aes.BlockSize)
	defer region.Free(pkt.SourceBufferAddress)

	pkt.DestinationBufferAddress = pkt.SourceBufferAddress

	pkt.PayloadPointer = region.Alloc(iv, 0)
	defer region.Free(pkt.PayloadPointer)

	ptr := region.Alloc(pkt.Bytes(), 0)
	defer region.Free(ptr)

	err = cmd(ptr, 1)

	if err != nil {
		return nil, err
	}

	if index >= 0 {
		err = setKeyData(index, nil, pkt.SourceBufferAddress)
	} else {
		region.Read(pkt.SourceBufferAddress, 0, key)
	}

	return
}

func setKeyData(index int, key []byte, addr uint32) (err error) {
	var keyLocation uint32
	var subword uint32

	if index < 0 || index > 3 {
		return errors.New("key index must be between 0 and 3")
	}

	if key != nil && len(key) > aes.BlockSize {
		return errors.New("invalid key size")
	}

	bits.SetN(&keyLocation, KEY_INDEX, 0b11, uint32(index))

	mux.Lock()
	defer mux.Unlock()

	for subword < 4 {
		off := subword * 4

		bits.SetN(&keyLocation, KEY_SUBWORD, 0b11, subword)
		reg.Write(DCP_KEY, keyLocation)

		if key != nil {
			k := key[off : off+4]
			reg.Write(DCP_KEYDATA, binary.LittleEndian.Uint32(k))
		} else {
			reg.Move(DCP_KEYDATA, addr+off)
		}

		subword++
	}

	return
}

// SetKey configures an AES-128 key in one of the 4 available slots of the DCP
// key RAM.
func SetKey(index int, key []byte) (err error) {
	return setKeyData(index, key, 0)
}
