package message

import (
	"errors"

	"../amf"
	"../conn"
	chunk "../ref-chunk"
	"github.com/sirupsen/logrus"
)

func cmdChunk(Fmt uint32, CSID uint32, msgStreamID uint32, msgSize uint32) *chunk.Chunk {
	cmdChunk := chunk.NewChunk()
	cmdChunk.BasicHeader.Fmt = Fmt
	cmdChunk.BasicHeader.CSID = CSID
	cmdChunk.MessageHeader.MessageTypeID = 20
	cmdChunk.MessageHeader.MessageLength = msgSize
	cmdChunk.MessageHeader.MessageStreamID = msgStreamID
	cmdChunk.MessageHeader.TtimeStamp = 0
	return cmdChunk
}

func writeMsg(conn *conn.Conn, CSID uint32, streamID uint32, args ...interface{}) error {
	conn.Bytesw.Reset()
	for _, v := range args {
		if _, err := conn.Encoder.Encode(conn.Bytesw, v, amf.AMF0); err != nil {
			return err
		}
	}
	msg := conn.Bytesw.Bytes()

	message := &Message{
		MessageTypeID:   20,
		MessageStreamID: streamID,
		MessageData:     msg,
		MessageLength:   uint32(len(msg)),
		ReaderWriter:    conn.ReaderWriter,
		chunkSize:       conn.ServerChunkSize,
		Chunks:          make(map[uint32]*chunk.Chunk),
	}

	logrus.Debugf("[Debug] message is %v", message)
	firstChunk := cmdChunk(uint32(0), CSID, streamID, uint32(len(msg)))
	message.CreateChunksBasedOnFirstChunkThenWrite(firstChunk)
	return conn.ReaderWriter.Flush()
}

func handleConnectCmd(amfMessageData []interface{}, conn *conn.Conn) error {
	for _, v := range amfMessageData {
		switch v.(type) {
		case string:
		case float64:
			id := int(v.(float64))
			if id != 1 {
				return errors.New("id is invalid")
			}
			conn.ConnInfo.TransactionID = id
			logrus.Debugf("[Debug] transactionID is %v", conn.ConnInfo.TransactionID)
		case amf.Object:
			logrus.Debug("[Debug] start to gather information from amf")
			obimap := v.(amf.Object)
			if app, ok := obimap["app"]; ok {
				conn.ConnInfo.App = app.(string)
			}
			if flashVer, ok := obimap["flashVer"]; ok {
				conn.ConnInfo.Flashver = flashVer.(string)
			}
			if tcurl, ok := obimap["tcUrl"]; ok {
				conn.ConnInfo.TcURL = tcurl.(string)
			}
			if encoding, ok := obimap["objectEncoding"]; ok {
				conn.ConnInfo.ObjectEncoding = int(encoding.(float64))
			}
		}
	}
	return nil
}

func responseToConnect(conn *conn.Conn, msg *Message) error {
	controlMsg := NewWindowAckSizeMessage(conn, 2500000)
	logrus.Debug("[Debug] sending window ack size message")
	controlMsg.WriteWithProvidedChunkList()

	controlMsg = NewSetPeerBandwidthMessage(conn, 2500000)
	logrus.Debug("[Debug] set peer bandwidth message")
	controlMsg.WriteWithProvidedChunkList()

	controlMsg = NewSetChunkSizeMessage(conn, 1024)
	logrus.Debug("[Debug] set chunk size message")
	controlMsg.WriteWithProvidedChunkList()
	conn.ServerChunkSize = 1024

	resp := make(amf.Object)
	resp["fmsVer"] = "FMS/3,0,1,123"
	resp["capabilities"] = 31

	event := make(amf.Object)
	event["level"] = "status"
	event["code"] = "NetConnection.Connect.Success"
	event["description"] = "Connection succeeded."
	event["objectEncoding"] = conn.ConnInfo.ObjectEncoding

	logrus.Debug("------resp and event in connect --------")
	logrus.Debugf("[Debug] resp is %v", resp)
	logrus.Debugf("[Debug] event is %v", event)
	logrus.Debug("------end of resp and event in connect --------")

	return writeMsg(conn, msg.Chunks[0].BasicHeader.CSID, msg.MessageStreamID, "_result", conn.ConnInfo.TransactionID, resp, event)
}

func handleCreateStreamCmd(amfMessageData []interface{}, conn *conn.Conn) error {
	for _, v := range amfMessageData {
		switch v.(type) {
		case string:
		case float64:
			conn.ConnInfo.TransactionID = int(v.(float64))
		case amf.Object:
		}
	}
	return nil
}

func responseToCreateStream(conn *conn.Conn, msg *Message) error {
	return writeMsg(conn, msg.Chunks[0].BasicHeader.CSID, msg.MessageStreamID, "_result", conn.ConnInfo.TransactionID, nil, 1)
}

func handlePublishCmd(amfMessageData []interface{}, conn *conn.Conn) error {
	for _, v := range amfMessageData {
		switch v.(type) {
		case string:
			break
		case float64:
			id := int(v.(float64))
			conn.ConnInfo.TransactionID = id
		case amf.Object:
		}
	}

	return nil
}

func responseToPublish(conn *conn.Conn, msg *Message) error {
	event := make(amf.Object)
	event["level"] = "status"
	event["code"] = "NetStream.Publish.Start"
	event["description"] = "Start publising."
	return writeMsg(conn, msg.Chunks[0].BasicHeader.CSID, msg.MessageStreamID, "onStatus", 0, nil, event)
}
