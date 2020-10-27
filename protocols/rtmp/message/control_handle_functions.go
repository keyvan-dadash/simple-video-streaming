package message

import (
	"encoding/binary"

	"github.com/sirupsen/logrus"

	"../conn"

	"../amf"
)

func setChunkSizeIDHandler(msg *Message, conn *conn.Conn) {
	conn.ClientChunkSize = binary.BigEndian.Uint32(msg.MessageData)
	logrus.Debugf("got setChunkSize control msg with chunksize %v", conn.ClientChunkSize)
}

func abortHandler(msg *Message, conn *conn.Conn) {
	logrus.Debug("dont need to process chunk stream id because got abort messaage")
}

func ackHandler(msg *Message, conn *conn.Conn) {
	logrus.Debug("got ack msg")
}

func userControlHandler(msg *Message, conn *conn.Conn) {
	logrus.Debug("got user msg")
}

func windowAckSizeHandler(msg *Message, conn *conn.Conn) {
	conn.ClientWindowAckSize = binary.BigEndian.Uint32(msg.MessageData)
	logrus.Debugf("got window ack size msg contoller with size %v", conn.ClientWindowAckSize)
}

func setPeerBandWidthHandler(msg *Message, conn *conn.Conn) {
	logrus.Debug("got set peer badwidth")
}

func audioMessageHandler(msg *Message, conn *conn.Conn) {
	logrus.Debug("got audio message")
}

func videoMessageHandler(msg *Message, conn *conn.Conn) {
	logrus.Debug("got video message")
}

func dataMessageAmf3Handler(msg *Message, conn *conn.Conn) {
	logrus.Debug("got amf3 data message")
}

func sharedObjectAmf3Handler(msg *Message, conn *conn.Conn) {
	logrus.Debug("got shared object amf3 message")
}

func amf3Handler(msg *Message, conn *conn.Conn) {
	conn.SetAmfVersion(amf.AMF3)
	logrus.Debug("got shared object amf3 message")
}

func dataMessageAmf0Handler(msg *Message, conn *conn.Conn) {
	logrus.Debug("got amf0 data message")
}

func sharedObjectAmf0Handler(msg *Message, conn *conn.Conn) {
	logrus.Debug("got shared object amf0 message")
}

func amf0Handler(msg *Message, conn *conn.Conn) {
	conn.SetAmfVersion(amf.AMF0)
	logrus.Debug("got shared object amf0 message")
}

func aggregateMessageHandler(msg *Message, conn *conn.Conn) {
	logrus.Debug("got aggregateMessage")
}
