package chunk

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"net"

	"github.com/sirupsen/logrus"

	uuid "github.com/satori/go.uuid"
)

var (
	setChunkSizeID       = uint32(1)
	abortMessageID       = uint32(2)
	ackID                = uint32(3)
	userControlMessageID = uint32(4)
	windowAckSizeID      = uint32(5)
	setPeerBandWidth     = uint32(6)
	audioMessageID       = uint32(8)
	videoMessageID       = uint32(9)
	dataMessageAmf3ID    = uint32(15)
	sharedObjectAmf3ID   = uint32(16)
	amf3ID               = uint32(17)
	dataMessageAmf0ID    = uint32(18)
	sharedObjectAmf0ID   = uint32(19)
	amf0ID               = uint32(20)
	aggregateMessageID   = uint32(22)
)

//Conn is Custom Conn to handel connection
type Conn struct {
	net.Conn
	id                  string
	chunkSize           uint32
	remoteChunkSize     uint32
	windowAckSize       uint32
	remoteWindowAckSize uint32
	received            uint32
	ackReceived         uint32
	rw                  *ReaderWriter
	chunks              map[uint32]Chunk
}

//NewConn create Conn struct based on buffersize and conn that passed
func NewConn(c net.Conn, bufferSize int) *Conn {
	id, _ := uuid.NewV4()
	return &Conn{
		id:                  id.String(),
		Conn:                c,
		chunkSize:           128,
		remoteChunkSize:     128,
		windowAckSize:       2500000,
		remoteWindowAckSize: 2500000,
		rw:                  NewReaderWriter(bufio.NewReadWriter(bufio.NewReader(c), bufio.NewWriter(c))),
		chunks:              make(map[uint32]Chunk),
	}
}

func (c *Conn) WriteChunk(chunk *Chunk) {

	numberOfChunks := (chunk.messageLength / c.chunkSize)

	sentSize := uint32(0)
	for i := uint32(0); i < numberOfChunks; i++ {

		if sentSize == chunk.messageLength {
			break
		}
		if i == uint32(0) {
			chunk.fmt = 0
		} else {
			chunk.fmt = 3
		}

		startPtr := c.chunkSize * i
		endPtr := startPtr + c.chunkSize
		if endPtr > uint32(len(chunk.data)) {
			endPtr = uint32(len(chunk.data))
		}

		buffer := chunk.data[startPtr:endPtr]

		c.rw.Write(buffer)

		sentSize += (endPtr - startPtr)
	}

}

func (c *Conn) ReadChunk(chunk *Chunk) {
	switch chunk.CSID {
	case 0:
		chunk.CSID = c.rw.ReadNByte(1) + 64
	case 1:
		chunk.CSID += c.rw.ReadNByte(2) + 64
	}
	logrus.Debugf("Got Chunk CSID: %v", chunk.CSID)

	switch chunk.fmt {
	case 0:
		chunk.timeStamp = c.rw.ReadNByte(3)
		logrus.Debugf("[Type0] timestamp: %v", chunk.timeStamp)
		chunk.messageLength = c.rw.ReadNByte(3)
		logrus.Debugf("[Type0] messagelength: %v", chunk.messageLength)
		chunk.messageTypeID = c.rw.ReadNByte(1)
		logrus.Debugf("[Type0] messagetypeid: %v", chunk.messageTypeID)
		chunk.messageStreamID = c.rw.ReadNByte(4)
		logrus.Debugf("[Type0] messagesteamid: %v", chunk.messageStreamID)
		if chunk.timeStamp == 0xffffff {
			chunk.haveExtendedTimeStamp = true
			chunk.timeStamp = c.rw.ReadNByte(4)
		} else {
			chunk.haveExtendedTimeStamp = false
		}

		chunk.data = make([]byte, chunk.messageLength)
		chunk.currentPos = 0
		chunk.isFinished = false
		logrus.Debug("finished chunk type 0")
	case 1:
		chunk.timeStampDelta = c.rw.ReadNByte(3)
		chunk.messageLength = c.rw.ReadNByte(3)
		chunk.messageTypeID = c.rw.ReadNByte(1)

		if chunk.timeStampDelta == 0xffffff {
			chunk.haveExtendedTimeStamp = true
			chunk.timeStampDelta = c.rw.ReadNByte(4)
		} else {
			chunk.haveExtendedTimeStamp = false
		}

		chunk.timeStamp += chunk.timeStampDelta
		logrus.Debug("finished chunk type 3")
	case 2:
		chunk.timeStampDelta = c.rw.ReadNByte(3)

		if chunk.timeStamp == 0xffffff {
			chunk.haveExtendedTimeStamp = true
			chunk.timeStampDelta = c.rw.ReadNByte(4)
		} else {
			chunk.haveExtendedTimeStamp = false
		}

		chunk.timeStamp += chunk.timeStampDelta
		logrus.Debug("finished chunk type 2")
	case 3:
		//i dont know
	}

	logrus.Debugf("Start create buffer and copy chunk len : %v, current pos: %v, chunk type id : %v, chunk fmt : %v, chunk message type : %v",
		chunk.messageLength, chunk.currentPos, chunk.CSID, chunk.fmt, chunk.messageTypeID)

	size := uint32(0)
	if chunk.currentPos+DefaultChunkSize > uint32(len(chunk.data)) {
		size = uint32(len(chunk.data))
	} else {
		size = chunk.currentPos + DefaultChunkSize
	}
	buffer := chunk.data[chunk.currentPos:size]

	logrus.Debug("start reading buffer")
	if _, err := c.rw.Read(buffer); err != nil {
		logrus.Errorf("[Error] error during reading and copy to buffer err: %v", err)
	}
	logrus.Debug("finished reading buffer")

	chunk.currentPos += uint32(len(buffer))
	if chunk.currentPos == chunk.messageLength {
		chunk.isFinished = true
	}

}

func (c *Conn) handleControlMessages(chunk *Chunk) {

	if chunk.messageTypeID == setChunkSizeID {
		fmt.Println(binary.BigEndian.Uint32(chunk.data))
		c.remoteChunkSize = binary.BigEndian.Uint32(chunk.data)
	} else if chunk.messageTypeID == windowAckSizeID {
		c.remoteWindowAckSize = binary.BigEndian.Uint32(chunk.data)
	}
	//can add extra message control here
}

func (c *Conn) Ack(chunk *Chunk) {
	c.ackReceived += chunk.messageLength
	if c.ackReceived > c.remoteWindowAckSize {
		ackChunk := CreateAck(*chunk)
		c.WriteChunk(ackChunk)
		c.ackReceived = 0
	}
}

func CreateAck(chunk Chunk) *Chunk {

	ackChunk := &Chunk{
		fmt:             0,
		CSID:            2,
		messageTypeID:   ackID,
		messageStreamID: 0,
		messageLength:   chunk.messageLength,
		data:            make([]byte, 4),
	}

	binary.BigEndian.PutUint32(ackChunk.data[:4], chunk.messageLength)

	return ackChunk

}
