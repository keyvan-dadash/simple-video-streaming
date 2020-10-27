package chunk

import (
	"bufio"
	"encoding/binary"

	"github.com/sirupsen/logrus"
)

//BasicHeader present of first header in chunk according to rtmp specification
type BasicHeader struct {
	Fmt  uint32
	CSID uint32
}

//NewBasicHeader return basicheader
func NewBasicHeader() *BasicHeader {
	return &BasicHeader{}
}

func (b *BasicHeader) Read(reader *bufio.Reader) (int, error) {
	recvBytes := 0
	basicHeader, err := reader.ReadByte()
	if err != nil {
		logrus.Errorf("[Error]Error occurred during read basic header first byte err: %v", err)
		return recvBytes, err
	}
	recvBytes++
	b.Fmt = uint32(basicHeader >> 6)
	tempCSID := uint32(basicHeader & 0x3f)
	if tempCSID == 0 {
		oneByteCsidMinus64, err := reader.ReadByte()
		if err != nil {
			logrus.Errorf("[Error]Error occurred during read basic header second byte header err: %v", err)
			return recvBytes, err
		}
		b.CSID = uint32((oneByteCsidMinus64 + 64))
		recvBytes++
	} else if tempCSID == 1 {
		twoByteCsidMinus64, err := ReadNByte(reader, 2)
		if err != nil {
			logrus.Errorf("[Error]Error occurred during read basic header second and third byte header err: %v", err)
			return recvBytes, err
		}
		b.CSID = uint32((twoByteCsidMinus64 + 64))
		recvBytes += 2
	} else {
		b.CSID = tempCSID
	}

	logrus.Debugf("[Debug] read basicheader with fmt %v and CSID %v", b.Fmt, b.CSID)
	return recvBytes, nil
}

func (b *BasicHeader) Write(writer *bufio.Writer) (int, error) {
	numberOfWrittenBytes := 0
	err := func() error { return nil }()
	if b.CSID < 64 {
		basicHeader := make([]byte, 1)
		basicHeader[0] = byte(uint8(b.Fmt) << 6)
		basicHeader[0] |= byte(uint8(b.CSID))
		numberOfWrittenBytes, err = writer.Write(basicHeader)
	} else if b.CSID-64 < 256 {
		basicHeader := make([]byte, 2)
		basicHeader[0] = byte(uint8(b.Fmt) << 6)
		basicHeader[0] |= 0
		basicHeader[1] = byte(uint8(b.CSID - 64))
		numberOfWrittenBytes, err = writer.Write(basicHeader)
	} else if b.CSID-64 < 65536 {
		basicHeader := make([]byte, 3)
		basicHeader[0] = byte(uint8(b.Fmt) << 6)
		basicHeader[0] |= 1
		binary.BigEndian.PutUint16(basicHeader[1:], uint16(b.CSID-64))
		numberOfWrittenBytes, err = writer.Write(basicHeader)
	} else {
		logrus.Errorf("[Error] CSID in chunk is invalid, CSID is %v", b.CSID)
	}
	if err != nil {
		logrus.Errorf("[Error] %v occurred during write basic header with csid %v and fmt %v", err, b.CSID, b.Fmt)
	}
	return numberOfWrittenBytes, err
}
