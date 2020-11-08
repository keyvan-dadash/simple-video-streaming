package chunk

import (
	"bufio"

	"github.com/sirupsen/logrus"
)

//Chunk is basic packet in rtmp protocol that transfer over network
type Chunk struct {
	BasicHeader       *BasicHeader
	MessageHeader     *MessageHeader
	ExtendedTimeStamp *ExtendedTimeStamp
}

//NewChunk return chunk
func NewChunk() *Chunk {
	return &Chunk{
		BasicHeader:       NewBasicHeader(),
		MessageHeader:     NewMessageHeader(),
		ExtendedTimeStamp: NewExtendedTimeStamp(),
	}
}

func (c *Chunk) Read(reader *bufio.Reader) error {
	logrus.Debug("[Debug] start reading chunk headers")
	if _, err := c.BasicHeader.Read(reader); err != nil {
		logrus.Errorf("[Error] Error during read basic header with err %v", err)
		return err
	}

	if _, err := c.MessageHeader.Read(reader, c.BasicHeader.Fmt); err != nil {
		logrus.Errorf("[Error] Error during read message header with err %v", err)
		return err
	}

	if c.MessageHeader.TtimeStamp == 0xffffff || c.MessageHeader.TimeStampDelta == 0xffffff {
		c.ExtendedTimeStamp.hasExtendedTimeStamp = true
	}

	if _, err := c.ExtendedTimeStamp.Read(reader); err != nil {
		logrus.Errorf("[Error] Error during read extended time stamp with err %v", err)
		return err
	}

	logrus.Debug("[Debug] finished reading chunk headers")
	return nil
}

//ReadPayload is function that read payload that in chunk
func (c *Chunk) ReadPayload(reader *bufio.Reader, chunkData []byte) error {
	if _, err := reader.Read(chunkData); err != nil {
		logrus.Errorf("[Error] %v occured during reading chunk with chunk stream ID %v", err, c.BasicHeader.CSID)
		return err
	}

	return nil
}

func (c *Chunk) Write(writer *bufio.Writer, chunkData []byte) error {
	logrus.Debug("[Debug] start writing chunk")
	if _, err := c.BasicHeader.Write(writer); err != nil {
		logrus.Errorf("[Error] Error during write basic header with err %v", err)
		return err
	}

	if _, err := c.MessageHeader.Write(writer, c.BasicHeader.Fmt); err != nil {
		logrus.Errorf("[Error] Error during write message header with err %v", err)
		return err
	}

	if c.MessageHeader.TtimeStamp == 0xffffff || c.MessageHeader.TimeStampDelta == 0xffffff {
		c.ExtendedTimeStamp.hasExtendedTimeStamp = true
	}

	if _, err := c.ExtendedTimeStamp.Write(writer); err != nil {
		logrus.Errorf("[Error] Error during write extended time stamp with err %v", err)
		return err
	}

	if _, err := writer.Write(chunkData); err != nil {
		logrus.Errorf("[Error] %v occured during writing chunk with chunk stream ID %v", err, c.BasicHeader.CSID)
		return err
	}

	logrus.Debug("[Debug] finished writing chunk")
	return nil
}
