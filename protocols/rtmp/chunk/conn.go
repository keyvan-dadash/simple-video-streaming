package chunk

import (
	"bufio"
	"log"
	"net"

	uuid "github.com/satori/go.uuid"
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

func (c *Conn) Read(chunk *Chunk) {

	for {
		header, err := c.rw.ReadByte()

		if err != nil {
			log.Panicf("Error in reading Header err: %v", err)
		}

		chunk.fmt = uint32(header >> 6)
		chunk.CSID = uint32(header & 0x3f)

		switch chunk.CSID {
		case 0:
			chunk.CSID = c.rw.ReadNByte(1) + 64
		case 1:
			chunk.CSID += c.rw.ReadNByte(2) + 64
		}

		switch chunk.fmt {
		case 0:
			chunk.timeStamp = c.rw.ReadNByte(3)
			chunk.messageLength = c.rw.ReadNByte(3)
			chunk.messageTypeID = c.rw.ReadNByte(1)
			chunk.messageStreamID = c.rw.ReadNByte(4)
			if chunk.timeStamp == 0xffffff {
				chunk.haveExtendedTimeStamp = true
				chunk.timeStamp = c.rw.ReadNByte(4)
			} else {
				chunk.haveExtendedTimeStamp = false
			}
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
		case 2:
			chunk.timeStampDelta = c.rw.ReadNByte(3)

			if chunk.timeStamp == 0xffffff {
				chunk.haveExtendedTimeStamp = true
				chunk.timeStampDelta = c.rw.ReadNByte(4)
			} else {
				chunk.haveExtendedTimeStamp = false
			}

			chunk.timeStamp += chunk.timeStampDelta
		case 3:
			//i dont know
		}

		cs, ok := c.chunks[chunk.CSID]
		if !ok {
			cs = Chunk{}
			c.chunks[chunk.CSID] = cs
		}

	}
}
