package hps

import "github.com/go-ble/ble"

// Private 128-bit UUIDs, which avoids the base of pre-defined 16/32-bits UUIDS
// xxxxxxxx-0000-1000-8000-00805F9B34FB [Vol 3, Part B, 2.5.1].
var (
	TestSvcUUID   = ble.MustParse("00010000-0001-1000-8000-00805F9B34FB")
	CountCharUUID = ble.MustParse("00010000-0002-1000-8000-00805F9B34FB")
	EchoCharUUID  = ble.MustParse("00020000-0002-1000-8000-00805F9B34FB")

	HpsServiceID = ble.MustParse("0136bd82-ba81-48c6-b608-df7aa274338a")

	URIUUID          = ble.MustParse("0x2AB6")
	ControlPointUUID = ble.MustParse("0x2ABA")
)

// const (
// 	PeripheralID = "b4a77f05-2524-4330-bcbb-5aafd2a9329b"
// 	DeviceName   = "davidoram/HPS"

// 	// From https://btprodspecificationrefs.blob.core.windows.net/assigned-values/16-bit%20UUID%20Numbers%20Document.pdf
// 	// HTTPURIID          = 0x2AB6
// 	HTTPHeadersID    = 0x2AB7
// 	HTTPStatusCodeID = 0x2AB8
// 	HTTPEntityBodyID = 0x2AB9

// 	HTTPSSecurityID   = 0x2ABB
// 	TDSControlPointID = 0x2ABC

// 	HTTPReserved      uint8 = 0x00
// 	HTTPGet           uint8 = 0x01
// 	HTTPHead          uint8 = 0x02
// 	HTTPPost          uint8 = 0x03
// 	HTTPPut           uint8 = 0x04
// 	HTTPDelete        uint8 = 0x05
// 	HTTPSGet          uint8 = 0x06
// 	HTTPSHead         uint8 = 0x07
// 	HTTPSPost         uint8 = 0x08
// 	HTTPSPut          uint8 = 0x09
// 	HTTPSDelete       uint8 = 0x0a
// 	HTTPRequestCancel uint8 = 0x0b

// 	// Encode these values together in one octet
// 	HeadersReceived  uint8 = 0x01
// 	HeadersTruncated uint8 = 0x02
// 	BodyReceived     uint8 = 0x04
// 	BodyTruncated    uint8 = 0x08

// 	// HeaderMaxOctets is max buffer size that the HTTP Headers encode into,
// 	// otherwise the server will report HeadersTruncated
// 	HeaderMaxOctets int = 512

// 	// BodyMaxOctets is max buffer size of the HTTP Body,
// 	// otherwise the server will report BodyTruncated
// 	BodyMaxOctets int = 512

// 	DataStatusHeadersReceived  uint8 = 0x01
// 	DataStatusHeadersTruncated uint8 = 0x02
// 	DataStatusBodyReceived     uint8 = 0x04
// 	DataStatusBodyTruncated    uint8 = 0x08
// )
