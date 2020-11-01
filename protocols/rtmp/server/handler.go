package server

import (
	"../conn"
	"../message"
	"github.com/sirupsen/logrus"
)

//ConnHandler is structure that handle Conn
type ConnHandler struct {
	Conn     *conn.Conn
	messages map[uint32]*message.Message
	index    uint32
}

//NewConnHadler return handler for handle Conn
func NewConnHadler(c *conn.Conn) *ConnHandler {
	return &ConnHandler{
		Conn:     c,
		messages: make(map[uint32]*message.Message),
		index:    0,
	}
}

//Handle the Conn
func (ch *ConnHandler) Handle() {
	handshakeHandler := conn.NewHandShakeConn(ch.Conn)
	handshakeHandler.HandShake()
	previousMsg := message.NewMessage(ch.Conn.ReaderWriter, ch.Conn.ClientChunkSize, ch.Conn)

	for {
		msg := message.NewMessage(ch.Conn.ReaderWriter, ch.Conn.ClientChunkSize, ch.Conn)
		msg.FetchWithPreviousMsg(previousMsg)
		msg.Read() //note: we must fetch message
		previousMsg = msg
		ch.messages[msg.MessageStreamID] = msg
		msgCtlHandler := message.NewMsgControlHandler(msg, ch.Conn)
		msgCtlHandler.HandleMsgControl()
		if msg.MessageTypeID == 17 || msg.MessageTypeID == 20 {
			msgCmdHandler := message.NewMsgCmdHandler(msg, ch.Conn)
			msgCmdHandler.HandleMsgCmd()
		}
		if ch.Conn.AckReceived >= 0xf0000000 {
			ch.Conn.AckReceived = 0
		}
		if ch.Conn.AckReceived >= ch.Conn.ClientWindowAckSize {
			sendAck := message.NewAckMessage(ch.Conn, ch.Conn.AckReceived)
			logrus.Debugf("[Debug] sending ack with size %v", ch.Conn.AckReceived)
			sendAck.Write()
			ch.Conn.AckReceived = 0
		}
		ch.index++
	}
}

//GetMessageWithStreamID return msg with given streamid
func (ch *ConnHandler) GetMessageWithStreamID(msgStreamID uint32) *message.Message {
	return ch.messages[msgStreamID]
}
