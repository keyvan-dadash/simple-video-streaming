package message

import (
	"bytes"
	"io"
	"os"

	"github.com/sirupsen/logrus"

	"../amf"
	"../conn"
)

var (
	cmdConnect       = "connect"
	cmdFcpublish     = "FCPublish"
	cmdReleaseStream = "releaseStream"
	cmdCreateStream  = "createStream"
	cmdPublish       = "publish"
	cmdFCUnpublish   = "FCUnpublish"
	cmdDeleteStream  = "deleteStream"
	cmdPlay          = "play"
)

//MsgCmdHandler is structure that has functionality of handle cmd message
type MsgCmdHandler struct {
	Msg  *Message
	conn *conn.Conn
}

//NewMsgCmdHandler return msgcmdhandler with given message
func NewMsgCmdHandler(msg *Message, conn *conn.Conn) *MsgCmdHandler {
	return &MsgCmdHandler{
		Msg:  msg,
		conn: conn,
	}
}

//HandleMsgCmd handle the msg command and use when messagetypeid is 17 or 20
func (msgCH *MsgCmdHandler) HandleMsgCmd() error {

	decodedMsg, err := msgCH.decodeAmfCommand()
	if err != nil {
		return err
	}

	logrus.Debugf("[Debug]decoded amf msg with value %v", decodedMsg)

	switch decodedMsg[0].(type) {
	case string:
		switch decodedMsg[0].(string) {
		case cmdConnect:
			if err := handleConnectCmd(decodedMsg[1:], msgCH.conn); err != nil {
				return err
			}

			if err := responseToConnect(msgCH.conn, msgCH.Msg); err != nil {
				return err
			}
		case cmdCreateStream:
			handleCreateStreamCmd(decodedMsg[1:], msgCH.conn)
			responseToCreateStream(msgCH.conn, msgCH.Msg)
			logrus.Debug("[Debug] finished create stream")
			// os.Exit(1)
		case cmdPlay:
			logrus.Debug("finished play")
			os.Exit(1)
		case cmdPublish:
			handlePublishCmd(decodedMsg[1:], msgCH.conn)
			responseToPublish(msgCH.conn, msgCH.Msg)
		case cmdFcpublish:
			// connServer.fcPublish(decodedMsg)
		case cmdReleaseStream:
			// connServer.releaseStream(decodedMsg)
		case cmdFCUnpublish:
		case cmdDeleteStream:
		default:
			logrus.Debug("no support command=", decodedMsg[0].(string))
		}
	}

	return nil
}

func (msgCH *MsgCmdHandler) decodeAmfCommand() ([]interface{}, error) {
	if msgCH.conn.GetAmfVersion() == amf.AMF3 {
		msgCH.Msg.MessageData = msgCH.Msg.MessageData[1:] //ignore first byte bcs its present version
	}

	reader := bytes.NewReader(msgCH.Msg.MessageData)
	amfVersion := msgCH.conn.GetAmfVersion()
	decodedCmd, err := msgCH.conn.Decoder.DecodeBatch(reader, amfVersion)
	if err != nil && err != io.EOF {
		logrus.Debugf("[Debug] error occured during decode amf command err: %v", err)
		return nil, err
	}

	return decodedCmd, nil
}
