package chunk

import (
	"log"

	"github.com/sirupsen/logrus"
)

func Handler(conn *Conn) {

	for {
		for {
			logrus.Debug("start get chunk from conn")
			conn.rw.Flush()
			header, err := conn.rw.ReadByte()

			if err != nil {
				log.Panicf("Error in reading Header err: %v", err)
			}

			fmt := uint32(header >> 6)
			csid := uint32(header & 0x3f)

			logrus.Debugf("[Debug] Got chunk with fmt: %v and csid: %v", fmt, csid)

			c, ok := conn.chunks[csid]
			if !ok {
				c = &Chunk{}
				conn.chunks[csid] = c
			}
			c.fmt = fmt
			c.CSID = csid
			conn.ReadChunk(c)
			if c.isFinished {
				conn.handleControlMessages(c)
				logrus.Debug("finished reading message start to send ack")
				logrus.Debug(c.messageTypeID)
				conn.Ack(c)
				if c.messageTypeID == 17 || c.messageTypeID == 20 {
					conn.HandleCmdMsg(c)
				}
				conn.rw.Flush()
				break
			}

		}
	}

}
