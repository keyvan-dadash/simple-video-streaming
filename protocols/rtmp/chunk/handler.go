package chunk

import (
	"log"

	"github.com/sirupsen/logrus"
)

func Handler(conn *Conn) {

	for {
		header, err := conn.rw.ReadByte()

		if err != nil {
			log.Panicf("Error in reading Header err: %v", err)
		}

		fmt := uint32(header >> 6)
		csid := uint32(header & 0x3f)

		logrus.Debugf("[Debug] Got chunk with fmt: %v and csid: %v", fmt, csid)

		c, ok := conn.chunks[csid]
		if !ok {
			c = Chunk{}
			conn.chunks[csid] = c
		}
		c.fmt = fmt
		c.CSID = csid
		conn.ReadChunk(&c)
		if c.isFinished {
			conn.handleControlMessages(&c)
			logrus.Debug("finished reading message start to send ack")
			conn.Ack(&c)
			conn.rw.Flush()
			logrus.Debug("ack sent successfully")
			break
		}

	}

	Handler(conn)

}
