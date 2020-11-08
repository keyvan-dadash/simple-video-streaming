package message

import (
	"bytes"
	"io"

	"github.com/sirupsen/logrus"

	"../amf"
	"../conn"
)

var (
	CmdConnect       = "connect"
	CmdFcpublish     = "FCPublish"
	CmdReleaseStream = "releaseStream"
	CmdCreateStream  = "createStream"
	CmdPublish       = "publish"
	CmdFCUnpublish   = "FCUnpublish"
	CmdDeleteStream  = "deleteStream"
	CmdPlay          = "play"

	CmdError = "error"
)

//MsgCmdHandler is structure that has functionality of handle cmd message
type MsgCmdHandler struct {
	Msg  *Message
	conn *conn.Conn
	Type string
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
		logrus.Errorf("[Error] Amf cannot decode because faced to error %v", err)
		return err
	}

	logrus.Debugf("[Debug]decoded amf msg with value %v", decodedMsg)

	switch decodedMsg[0].(type) {
	case string:
		switch decodedMsg[0].(string) {
		case CmdConnect:
			if err := handleConnectCmd(decodedMsg[1:], msgCH.conn); err != nil {
				logrus.Errorf("[Error] Handle Connect Cmd faced to error %v", err)
				return err
			}

			if err := responseToConnect(msgCH.conn, msgCH.Msg); err != nil {
				logrus.Errorf("[Error] Response to Connect Command faced to error %v", err)
				return err
			}
		case CmdCreateStream:
			if err := handleCreateStreamCmd(decodedMsg[1:], msgCH.conn); err != nil {
				logrus.Errorf("[Error] Handle Create Stream Command faced to error %v", err)
				return err
			}

			if err := responseToCreateStream(msgCH.conn, msgCH.Msg); err != nil {
				logrus.Errorf("[Error] Response to Create Stream Command faced to error %v",
					err)
				return err
			}

			msgCH.Type = CmdCreateStream
			logrus.Debug("[Debug] finished create stream")
			// os.Exit(1)
		case CmdPlay:
			logrus.Debug("finished play")
			msgCH.Type = CmdPlay
		case CmdPublish:
			if err := handlePublishCmd(decodedMsg[1:], msgCH.conn); err != nil {
				logrus.Errorf("[Error] Handle Publish Command faced to error %v", err)
				return err
			}

			if err := responseToPublish(msgCH.conn, msgCH.Msg); err != nil {
				logrus.Errorf("[Error] Response to Publish Command faced to error %v",
					err)
				return err
			}

			msgCH.Type = CmdPublish
		case CmdFcpublish:
			// connServer.fcPublish(decodedMsg)
			msgCH.Type = CmdFcpublish
		case CmdReleaseStream:
			// connServer.releaseStream(decodedMsg)
		case CmdFCUnpublish:
			msgCH.Type = CmdFCUnpublish
		case CmdDeleteStream:
			msgCH.Type = CmdDeleteStream
		default:
			logrus.Debug("no support command=", decodedMsg[0].(string))
			msgCH.Type = CmdError
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
		logrus.Debugf("[Debug] Error occured during decode amf command err: %v", err)
		return nil, err
	}

	return decodedCmd, nil
}
