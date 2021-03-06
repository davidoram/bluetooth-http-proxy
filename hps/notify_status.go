package hps

import (
	"bytes"
	"encoding/binary"
	"log"
)

type NotifyStatus struct {
	HeadersReceived  bool
	HeadersTruncated bool
	BodyReceived     bool
	BodyTruncated    bool
	StatusCode       int
}

const (
	// Encode these values together in one octet
	HeadersReceived  uint8 = 0x01
	HeadersTruncated uint8 = 0x02
	BodyReceived     uint8 = 0x04
	BodyTruncated    uint8 = 0x08
)

func (n NotifyStatus) Encode() []byte {
	var dataStatus uint8
	if n.BodyReceived {
		dataStatus = dataStatus | BodyReceived
	}
	if n.BodyTruncated {
		dataStatus = dataStatus | BodyTruncated
	}
	if n.HeadersReceived {
		dataStatus = dataStatus | HeadersReceived
	}
	if n.HeadersTruncated {
		dataStatus = dataStatus | HeadersTruncated
	}

	// Http Status code eg: 200
	var sc uint16 = uint16(n.StatusCode)

	var b bytes.Buffer
	binary.Write(&b, binary.LittleEndian, sc)
	binary.Write(&b, binary.LittleEndian, dataStatus)
	return b.Bytes()
}

func DecodeNotifyStatus(buf []byte) (NotifyStatus, error) {

	var ns NotifyStatus
	r := bytes.NewReader(buf)

	var data struct {
		StatusCode uint16
		DataStatus uint8
	}

	if err := binary.Read(r, binary.LittleEndian, &data); err != nil {
		log.Println("binary.Read failed:", err)
		return ns, err
	}
	ns.HeadersReceived = data.DataStatus&HeadersReceived == HeadersReceived
	ns.HeadersTruncated = data.DataStatus&HeadersTruncated == HeadersTruncated
	ns.BodyReceived = data.DataStatus&BodyReceived == BodyReceived
	ns.BodyTruncated = data.DataStatus&BodyTruncated == BodyTruncated

	// Http Status code eg: 200
	ns.StatusCode = int(data.StatusCode)

	return ns, nil
}
