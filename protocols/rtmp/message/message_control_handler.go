package message

import (
	"../conn"
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

//MsgControlHandler is structure that has functionality of controling message
type MsgControlHandler struct {
	Msg  *Message
	conn *conn.Conn
}

//NewMsgControlHandler return msgcontolhandler with given message
func NewMsgControlHandler(msg *Message, conn *conn.Conn) *MsgControlHandler {
	return &MsgControlHandler{
		Msg:  msg,
		conn: conn,
	}
}

//HandleMsgControl is function that handle contol part of msg
func (msgCH *MsgControlHandler) HandleMsgControl() error {
	if msgCH.Msg.MessageStreamID == 0 && msgCH.Msg.Chunks[0].BasicHeader.CSID == 2 {
		switch msgCH.Msg.MessageTypeID {
		case setChunkSizeID:
			setChunkSizeIDHandler(msgCH.Msg, msgCH.conn)
		case abortMessageID:
			abortHandler(msgCH.Msg, msgCH.conn)
		case ackID:
			ackHandler(msgCH.Msg, msgCH.conn)
		case userControlMessageID:
			userControlHandler(msgCH.Msg, msgCH.conn)
		case windowAckSizeID:
			windowAckSizeHandler(msgCH.Msg, msgCH.conn)
		case setPeerBandWidth:
			setPeerBandWidthHandler(msgCH.Msg, msgCH.conn)
		case audioMessageID:
			audioMessageHandler(msgCH.Msg, msgCH.conn)
		case videoMessageID:
			videoMessageHandler(msgCH.Msg, msgCH.conn)
		case dataMessageAmf3ID:
			dataMessageAmf3Handler(msgCH.Msg, msgCH.conn)
		case sharedObjectAmf3ID:
			sharedObjectAmf3Handler(msgCH.Msg, msgCH.conn)
		case amf3ID:
			amf3Handler(msgCH.Msg, msgCH.conn)
		case dataMessageAmf0ID:
			dataMessageAmf0Handler(msgCH.Msg, msgCH.conn)
		case sharedObjectAmf0ID:
			sharedObjectAmf0Handler(msgCH.Msg, msgCH.conn)
		case amf0ID:
			amf0Handler(msgCH.Msg, msgCH.conn)
		case aggregateMessageID:
			aggregateMessageHandler(msgCH.Msg, msgCH.conn)
		}
	}
	return nil
}
