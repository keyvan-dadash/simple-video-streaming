package chunk

import (
	"bufio"
	"log"
	"net"
)

//Conn is Custom Conn to handel connection
type Conn struct {
	net.Conn
	chunkSize           uint32
	remoteChunkSize     uint32
	windowAckSize       uint32
	remoteWindowAckSize uint32
	received            uint32
	ackReceived         uint32
	rw                  *bufio.ReadWriter
	chunks              map[uint32]Chunk
}

//NewConn create Conn struct based on buffersize and conn that passed
func NewConn(c net.Conn, bufferSize int) *Conn {
	return &Conn{
		Conn:                c,
		chunkSize:           128,
		remoteChunkSize:     128,
		windowAckSize:       2500000,
		remoteWindowAckSize: 2500000,
		rw:                  bufio.NewReadWriter(bufio.NewReader(c), bufio.NewWriter(c)),
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

		cs, ok := c.chunks[chunk.CSID]
		if !ok {
			cs = Chunk{}
			c.chunks[chunk.CSID] = cs
		}

	}
}
