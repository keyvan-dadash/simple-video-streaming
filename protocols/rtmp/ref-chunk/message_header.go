package chunk

import (
	"bufio"
	"encoding/binary"

	"github.com/sirupsen/logrus"
)

var (
	TIMESTAMP_LENGTH       = 3
	MESSAGELENGTH_LENGTH   = 3
	MESSAGETYPEID_LENGTH   = 1
	MESSAGESTREAMID_LENGTH = 4
	TIMESTAMPDELTA_LENGTH  = 3
)

//MessageHeader present of second header in chunk according to rtmp specification
type MessageHeader struct {
	TtimeStamp      uint32
	MessageLength   uint32
	MessageTypeID   uint32
	MessageStreamID uint32
	TimeStampDelta  uint32
}

//NewMessageHeader return messageheader
func NewMessageHeader() *MessageHeader {
	return &MessageHeader{}
}

func (m *MessageHeader) Read(reader *bufio.Reader, fmt uint32) (int, error) {
	numberOfRecvByte := 0
	err := func() error { return nil }()
	switch fmt {
	case 0:
		m.TtimeStamp, err = ReadNByte(reader, TIMESTAMP_LENGTH)
		numberOfRecvByte += TIMESTAMP_LENGTH

		m.MessageLength, err = ReadNByte(reader, MESSAGELENGTH_LENGTH)
		numberOfRecvByte += MESSAGELENGTH_LENGTH

		m.MessageTypeID, err = ReadNByte(reader, MESSAGETYPEID_LENGTH)
		numberOfRecvByte += MESSAGETYPEID_LENGTH

		m.MessageStreamID, err = ReadNByte(reader, MESSAGESTREAMID_LENGTH)
		numberOfRecvByte += MESSAGESTREAMID_LENGTH

		logrus.Debugf("[Debug] read messageheader with fmt %v TimeStamp %v MessageLength %v MessageTypeID %v MessageStreamID %v",
			fmt, m.TtimeStamp, m.MessageLength, m.MessageTypeID, m.MessageStreamID)
	case 1:
		m.TimeStampDelta, err = ReadNByte(reader, TIMESTAMPDELTA_LENGTH)
		numberOfRecvByte += TIMESTAMPDELTA_LENGTH

		m.MessageLength, err = ReadNByte(reader, MESSAGELENGTH_LENGTH)
		numberOfRecvByte += MESSAGELENGTH_LENGTH

		m.MessageTypeID, err = ReadNByte(reader, MESSAGETYPEID_LENGTH)
		numberOfRecvByte += MESSAGETYPEID_LENGTH

		logrus.Debugf("[Debug] read messageheader with fmt %v TimeStampDelta %v MessageLength %v MessageTypeID %v",
			fmt, m.TimeStampDelta, m.MessageLength, m.MessageTypeID)
	case 2:
		m.TimeStampDelta, err = ReadNByte(reader, TIMESTAMPDELTA_LENGTH)
		numberOfRecvByte += TIMESTAMPDELTA_LENGTH

		logrus.Debugf("[Debug] read messageheader with fmt %v TimeStampDelta %v",
			fmt, m.TimeStampDelta)
	case 3:
		// ReadNByte(reader, TIMESTAMPDELTA_LENGTH+1)
		//we must determine this to how exact we are going to handle this
	}

	if err != nil {
		logrus.Errorf("[Error] %v occured during read message header with fmt %v", err, fmt)
	}

	return numberOfRecvByte, err
}

func (m *MessageHeader) Write(writer *bufio.Writer, fmt uint32) (int, error) {
	numberOfWrittenBytes := 0
	err := func() error { return nil }()
	switch fmt {
	case 0:
		nn := 0
		timeStamp := make([]byte, TIMESTAMP_LENGTH+1)
		binary.BigEndian.PutUint32(timeStamp, m.TtimeStamp)
		nn, err = writer.Write(timeStamp[1:])

		numberOfWrittenBytes += nn
		nn = 0

		messageLength := make([]byte, MESSAGELENGTH_LENGTH+1)
		binary.BigEndian.PutUint32(messageLength, m.MessageLength)
		nn, err = writer.Write(messageLength[1:])

		numberOfWrittenBytes += nn
		nn = 0

		messageTypeID := make([]byte, MESSAGETYPEID_LENGTH)
		messageTypeID[0] = uint8(m.MessageTypeID)
		nn, err = writer.Write(messageTypeID)

		numberOfWrittenBytes += nn
		nn = 0

		messageStreamID := make([]byte, MESSAGESTREAMID_LENGTH)
		binary.BigEndian.PutUint32(messageStreamID, m.MessageStreamID)
		nn, err = writer.Write(messageStreamID)

		numberOfWrittenBytes += nn
		nn = 0

		logrus.Debugf("[Debug] write messageheader with fmt %v TimeStamp %v MessageLength %v MessageTypeID %v MessageStreamID %v",
			fmt, timeStamp[1:], messageLength, messageTypeID, messageStreamID)

	case 1:
		nn := 0
		timeDelta := make([]byte, TIMESTAMPDELTA_LENGTH+1)
		binary.BigEndian.PutUint32(timeDelta, m.TimeStampDelta)
		nn, err = writer.Write(timeDelta[1:])

		numberOfWrittenBytes += nn
		nn = 0

		messageLength := make([]byte, MESSAGELENGTH_LENGTH+1)
		binary.BigEndian.PutUint32(messageLength, m.MessageLength)
		nn, err = writer.Write(messageLength[1:])

		numberOfWrittenBytes += nn
		nn = 0

		messageTypeID := make([]byte, MESSAGETYPEID_LENGTH)
		messageTypeID[0] = uint8(m.MessageTypeID)
		nn, err = writer.Write(messageTypeID)

		numberOfWrittenBytes += nn
		nn = 0

		logrus.Debugf("[Debug] write messageheader with fmt %v TimeStampDelta %v MessageLength %v MessageTypeID %v",
			fmt, m.TimeStampDelta, m.MessageLength, m.MessageTypeID)

	case 2:
		nn := 0
		timeDelta := make([]byte, TIMESTAMPDELTA_LENGTH)
		binary.BigEndian.PutUint32(timeDelta, m.TimeStampDelta)
		nn, err = writer.Write(timeDelta[1:])

		numberOfWrittenBytes += nn
		nn = 0

		logrus.Debugf("[Debug] write messageheader with fmt %v TimeStampDelta %v",
			fmt, m.TimeStampDelta)
	}

	if err != nil {
		logrus.Errorf("[Error] %v occured during write message header with fmt %v", err, fmt)
	}

	return numberOfWrittenBytes, err
}
