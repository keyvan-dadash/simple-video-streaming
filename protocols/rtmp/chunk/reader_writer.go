package chunk

import (
	"bufio"

	"github.com/sirupsen/logrus"
)

//ReaderWriter is custom ReaderWriter
type ReaderWriter struct {
	*bufio.ReadWriter
}

//NewReaderWriter create ReaderWriter and return pointer
func NewReaderWriter(rw *bufio.ReadWriter) *ReaderWriter {
	return &ReaderWriter{
		&bufio.ReadWriter{
			Reader: rw.Reader,
			Writer: rw.Writer,
		},
	}
}

//ReadNByte Read N Byte form readerwriter
func (rw *ReaderWriter) ReadNByte(n int) uint32 {
	if n < 0 {
		return 0
	}
	ret := uint32(0)
	for i := 0; i < n; i++ {
		by, err := rw.ReadByte()

		if err != nil {
			logrus.Errorf("[Error]Read %v Byte from readerwriter faced to error, err: %v", n, err)
		}
		ret = ret<<8 + uint32(by)
	}
	return ret
}
