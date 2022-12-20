package usb

import (
	"bytes"
	"encoding/binary"
)

const (
	HID_DESCRIPTOR_LENGTH = 0x09
	KEYBOARD_INTERFACE    = 0x21
)

// HIDDescriptor implements
// p22, Section 6.2.1 HID Descriptor
type HIDDescriptor struct {
	Length                 uint8
	DescriptorType         uint8
	bcdHID                 uint16
	CountryCode            uint8
	NumDescriptors         uint8
	ReportDescriptorType   uint8
	ReportDescriptorLength uint16
}

func (d *HIDDescriptor) SetKeyboardDefaults() {
	d.Length = HID_DESCRIPTOR_LENGTH
	d.DescriptorType = KEYBOARD_INTERFACE
	d.bcdHID = 0x101
	d.CountryCode = 33   // United States
	d.NumDescriptors = 1 // At least one for the report descriptor
	d.ReportDescriptorType = 0x22
	d.ReportDescriptorLength = 0x40
}

func (d *HIDDescriptor) Bytes() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, d)
	return buf.Bytes()
}

type HIDReportDescriptor []byte

// CoolermasterTKLSReportDescriptor returns bytes ripped from a coolermaster
// keyboard I had lying around.
func CoolermasterTKLSReportDescriptor() []byte {
	return []byte{
		0x05, 0x01, 0x09, 0x06, 0xa1, 0x01, 0x05, 0x07, 0x19, 0xe0, 0x29, 0xe7,
		0x15, 0x00, 0x25, 0x01, 0x75, 0x01, 0x95, 0x08, 0x81, 0x02, 0x95, 0x01,
		0x75, 0x08, 0x81, 0x03, 0x95, 0x03, 0x75, 0x01, 0x05, 0x08, 0x19, 0x01,
		0x29, 0x03, 0x91, 0x02, 0x95, 0x01, 0x75, 0x05, 0x91, 0x03, 0x95, 0x06,
		0x75, 0x08, 0x15, 0x00, 0x26, 0xa4, 0x00, 0x05, 0x07, 0x19, 0x00, 0x29,
		0xa4, 0x81, 0x00, 0xc0,
	}
}
