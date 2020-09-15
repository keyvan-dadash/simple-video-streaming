package chunk

import (
	"encoding/binary"

	"github.com/sirupsen/logrus"
)

//Chunk is base packet that send over rtmp protocol and have
//2 field and first field is Chunk header second is Chunk data
//Chunk header consist of 3 spereate header:
//1.basic header
//2.message header
//3.extended timestamp

var (

	//DefaultChunkSize is defualt for lenght of each chunk accord to rtmp specification
	DefaultChunkSize = uint32(128)
)

//Chunk is basic packet that transfer over network
type Chunk struct {

	//basic header
	fmt  uint32
	CSID uint32

	//message header
	timeStamp       uint32
	messageLength   uint32
	messageTypeID   uint32
	messageStreamID uint32
	timeStampDelta  uint32

	haveExtendedTimeStamp bool

	currentPos uint32
	isFinished bool
	//chunk data
	data []byte
}

func (c *Chunk) writeHeader(rw *ReaderWriter) {

	//write basic header
	if c.CSID < 64 {
		basicHeader := make([]byte, 1)
		basicHeader[0] = byte(uint8(c.fmt) << 6)
		basicHeader[0] |= byte(uint8(c.CSID))
		rw.Write(basicHeader)
	} else if c.CSID-64 < 256 {
		basicHeader := make([]byte, 2)
		basicHeader[0] = byte(uint8(c.fmt) << 6)
		basicHeader[0] |= 0
		basicHeader[1] = byte(uint8(c.CSID - 64))
		rw.Write(basicHeader)
	} else if c.CSID-64 < 65536 {
		basicHeader := make([]byte, 3)
		basicHeader[0] = byte(uint8(c.fmt) << 6)
		basicHeader[0] |= 1
		binary.BigEndian.PutUint16(basicHeader[1:], uint16(c.CSID-64))
		rw.Write(basicHeader)
	} else {
		logrus.Errorf("[Error] CSID in chunk is invalid, CSID is %v", c.CSID)
	}

	//message header
	switch c.fmt {
	case 0:
		timestamp := make([]byte, 4)
		if !c.haveExtendedTimeStamp {
			binary.BigEndian.PutUint32(timestamp, uint32(0xffffff))
			rw.Write(timestamp[1:])
		} else {
			binary.BigEndian.PutUint32(timestamp, c.timeStamp)
			rw.Write(timestamp[1:])
		}
		messageLength := make([]byte, 4)
		binary.BigEndian.PutUint32(messageLength, c.messageLength)
		rw.Write(messageLength[1:])

		messagetypeID := make([]byte, 1)
		messagetypeID[0] = uint8(c.messageTypeID)
		rw.Write(messagetypeID)

		messageStreamID := make([]byte, 4)
		binary.BigEndian.PutUint32(messageStreamID, c.messageStreamID)
		rw.Write(messageStreamID)
	case 1:
		timedelta := make([]byte, 4)
		if !c.haveExtendedTimeStamp {
			binary.BigEndian.PutUint32(timedelta, uint32(0xffffff))
			rw.Write(timedelta[1:])
		} else {
			binary.BigEndian.PutUint32(timedelta, c.timeStamp)
			rw.Write(timedelta[1:])
		}

		messageLength := make([]byte, 4)
		binary.BigEndian.PutUint32(messageLength, c.messageLength)
		rw.Write(messageLength[1:])

		messagetypeID := make([]byte, 1)
		messagetypeID[0] = uint8(c.messageTypeID)
		rw.Write(messagetypeID)
	case 2:
		timedelta := make([]byte, 4)
		if !c.haveExtendedTimeStamp {
			binary.BigEndian.PutUint32(timedelta, uint32(0xffffff))
			rw.Write(timedelta[1:])
		} else {
			binary.BigEndian.PutUint32(timedelta, c.timeStamp)
			rw.Write(timedelta[1:])
		}
	}

	if c.haveExtendedTimeStamp {
		timestamp := make([]byte, 4)
		binary.BigEndian.PutUint32(timestamp, c.timeStamp)
		rw.Write(timestamp)
	}
}
