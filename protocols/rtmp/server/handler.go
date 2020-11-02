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

//Handle is function that handle bood connection phase and media transfer phase
func (ch *ConnHandler) Handle() {
	ch.HandleMakingConnetionPhase()
	ch.HandleMediaTransferPhase()
}

//HandleMakingConnetionPhase is function that handle making phase of Connection
func (ch *ConnHandler) HandleMakingConnetionPhase() {
	ch.handShake()
	if err := ch.loopAndHandleMessages(); err != nil {
		ch.CloseConn()
	}
}

//HandleMediaTransferPhase is function that handle media transfer phase
func (ch *ConnHandler) HandleMediaTransferPhase() {
	ch.loopAndHandleMessages()
}

//CloseConn is function that close connection
func (ch *ConnHandler) CloseConn() {
	if err := ch.Conn.Close(); err != nil {

	}
}

func (ch *ConnHandler) loopAndHandleMessages() error {
	previousMsg := message.NewMessage(ch.Conn.ReaderWriter, ch.Conn.ClientChunkSize, ch.Conn)

	for {
		msg := message.NewMessage(ch.Conn.ReaderWriter, ch.Conn.ClientChunkSize, ch.Conn)
		msg.FetchWithPreviousMsg(previousMsg)
		if err := msg.Read(); err != nil {
			return err
		}
		previousMsg = msg
		ch.messages[msg.MessageStreamID] = msg
		msgCtlHandler := message.NewMsgControlHandler(msg, ch.Conn)
		msgCtlHandler.HandleMsgControl()
		if msg.MessageTypeID == 17 || msg.MessageTypeID == 20 {
			msgCmdHandler := message.NewMsgCmdHandler(msg, ch.Conn)
			msgCmdHandler.HandleMsgCmd()
		}
		ch.ack() //error handeling
		ch.index++
	}
}

func (ch *ConnHandler) handShake() {
	handshakeHandler := conn.NewHandShakeConn(ch.Conn)
	handshakeHandler.HandShake()
}

func (ch *ConnHandler) ack() error {
	if ch.Conn.AckReceived >= 0xf0000000 {
		ch.Conn.AckReceived = 0
	}
	if ch.Conn.AckReceived >= ch.Conn.ClientWindowAckSize {
		sendAck := message.NewAckMessage(ch.Conn, ch.Conn.AckReceived)
		logrus.Debugf("[Debug] sending ack with size %v", ch.Conn.AckReceived)
		if err := sendAck.WriteWithProvidedChunkList(); err != nil {
			return err
		}
		ch.Conn.AckReceived = 0
	}
	return nil
}

//GetMessageWithStreamID return msg with given streamid
func (ch *ConnHandler) GetMessageWithStreamID(msgStreamID uint32) *message.Message {
	return ch.messages[msgStreamID]
}
