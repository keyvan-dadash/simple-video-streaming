package chunk

import (
	"bufio"

	"github.com/sirupsen/logrus"
)

//ReadNByte is function that take numberofbytes to read from reader
func ReadNByte(reader *bufio.Reader, numberOfBytes int) (uint32, error) {
	if numberOfBytes < 0 {
		return 0, nil
	}
	result := uint32(0)
	for i := 0; i < numberOfBytes; i++ {
		byte, err := reader.ReadByte()

		if err != nil {
			logrus.Errorf("[Error]Read %v Byte from readerwriter faced to error, err: %v", numberOfBytes, err)
			return result, err
		}
		result = result<<8 + uint32(byte)
	}
	return result, nil
}
