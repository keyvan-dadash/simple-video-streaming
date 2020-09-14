package chunk

import (
	"bufio"

	"github.com/sirupsen/logrus"
)

type ReaderWriter struct {
	*bufio.ReadWriter
}

func NewReaderWriter(rw *bufio.ReadWriter) *ReaderWriter {
	return &ReaderWriter{
		&bufio.ReadWriter{
			Reader: rw.Reader,
			Writer: rw.Writer,
		},
	}
}
func (rw *ReaderWriter) ReadNByte(n int) uint32 {
	if n < 0 {
		return 0
	}
	ret := uint32(0)
	for i := 0; i < n; i++ {
		by, err := rw.ReadByte()

		if err != nil {
			logrus.Error("[Error]Read %v Byte from %rw faced to error", n, rw)
		}
		ret = ret<<8 + uint32(by)
	}
	return ret
}
