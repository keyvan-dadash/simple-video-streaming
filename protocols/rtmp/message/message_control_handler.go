package message

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

//MsgControlHandler is structure that has functionality os controling message
type MsgControlHandler struct {
	Msg *Message
}

//NewMsgControlHandler return msgcontolhandler with given message
func NewMsgControlHandler(msg *Message) *MsgControlHandler {
	return &MsgControlHandler{
		Msg: msg,
	}
}

//HandleMsg is function that handle contol part of msg
func (msgCH *MsgControlHandler) HandleMsg() error {
	if msgCH.Msg.MessageStreamID == 0 && msgCH.Msg.Chunks[0].BasicHeader.CSID == 2 {
		switch msgCH.Msg.MessageTypeID {
		case setChunkSizeID:
		case abortMessageID:
		case ackID:
		case userControlMessageID:
		case windowAckSizeID:
		case setPeerBandWidth:
		case audioMessageID:
		case videoMessageID:
		case dataMessageAmf3ID:
		case sharedObjectAmf3ID:
		case amf3ID:
		case dataMessageAmf0ID:
		case sharedObjectAmf0ID:
		case amf0ID:
		case aggregateMessageID:
		}
	}
	return nil
}
