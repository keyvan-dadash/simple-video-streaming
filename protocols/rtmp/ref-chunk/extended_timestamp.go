package chunk

import (
	"bufio"
	"encoding/binary"

	"github.com/sirupsen/logrus"
)

var (
	EXTENDEDTIMESTAMP_LENGTH = 4
)

//ExtendedTimeStamp present of third header(if exist) in chunk according to rtmp specification
type ExtendedTimeStamp struct {
	hasExtendedTimeStamp bool
	TimeStamp            uint32
}

//NewExtendedTimeStamp return extendedtimestamp
func NewExtendedTimeStamp() *ExtendedTimeStamp {
	return &ExtendedTimeStamp{
		hasExtendedTimeStamp: false,
	}
}

func (e *ExtendedTimeStamp) Read(reader *bufio.Reader) (int, error) {
	numberOfRecvBytes := 0
	err := func() error { return nil }()
	if e.hasExtendedTimeStamp {
		e.TimeStamp, err = ReadNByte(reader, EXTENDEDTIMESTAMP_LENGTH)
		numberOfRecvBytes += EXTENDEDTIMESTAMP_LENGTH
		logrus.Debugf("[Debug] read extended timestamp with value %v", e.TimeStamp)
	}

	if err != nil {
		logrus.Errorf("[Error] %v occured during read extendedtimestamp", err)
	}

	return numberOfRecvBytes, err
}

func (e *ExtendedTimeStamp) Write(writer *bufio.Writer) (int, error) {
	numberOfWrittenBytes := 0
	err := func() error { return nil }()
	if e.hasExtendedTimeStamp {
		extendedTimeStamp := make([]byte, 4)
		binary.BigEndian.PutUint32(extendedTimeStamp, e.TimeStamp)
		numberOfWrittenBytes, err = writer.Write(extendedTimeStamp)
		logrus.Debugf("[Debug] write extended timestamp with value %v", e.TimeStamp)
	}

	if err != nil {
		logrus.Errorf("[Error] %v occured during write extendedtimestamp", err)
	}

	return numberOfWrittenBytes, err
}
