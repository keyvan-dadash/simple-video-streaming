package message

import (
	"encoding/binary"

	"../conn"
	chunk "../ref-chunk"
)

func initControlChunk(messageTypeID uint32, msg *Message) *chunk.Chunk {

	controlChunk := chunk.NewChunk()
	controlChunk.BasicHeader.Fmt = 0
	controlChunk.BasicHeader.CSID = 2
	controlChunk.MessageHeader.MessageTypeID = messageTypeID
	controlChunk.MessageHeader.MessageLength = msg.MessageLength
	controlChunk.MessageHeader.MessageStreamID = msg.MessageStreamID

	return controlChunk
}

func initControlMessage(conn *conn.Conn, controlMessageValue uint32, ControlMessageID uint32, msgDataSize uint32) *Message {
	res := &Message{
		MessageStreamID:    0,
		MessageLength:      msgDataSize,
		ReaderWriter:       conn.ReaderWriter,
		index:              0,
		currentDataCellPos: 0,
		chunkSize:          conn.ClientChunkSize,
		MessageData:        make([]byte, msgDataSize),
		Chunks:             make(map[uint32]*chunk.Chunk),
	}

	binary.BigEndian.PutUint32(res.MessageData, controlMessageValue)

	controlChunk := initControlChunk(ControlMessageID, res)
	res.Chunks[res.index] = controlChunk

	return res
}

//NewWindowAckSizeMessage return WindowAckSize message with given conn
func NewWindowAckSizeMessage(conn *conn.Conn, windowAckSize uint32) *Message {
	return initControlMessage(conn, windowAckSize, windowAckSizeID, 4)
}

//NewSetPeerBandwidthMessage return setpeerbandwidth message with given conn
func NewSetPeerBandwidthMessage(conn *conn.Conn, bandwidth uint32) *Message {
	peerMsg := initControlMessage(conn, bandwidth, setPeerBandWidth, 5)
	peerMsg.MessageData[4] = 2
	return peerMsg
}

//NewSetChunkSizeMessage return setchunksize message with given conn
func NewSetChunkSizeMessage(conn *conn.Conn, chunkSize uint32) *Message {
	return initControlMessage(conn, chunkSize, setChunkSizeID, 4)
}

//NewAckMessage create Ack Size
func NewAckMessage(conn *conn.Conn, ackSize uint32) *Message {
	return initControlMessage(conn, ackSize, ackID, 4)
}
