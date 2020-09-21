package amf

import (
	"bytes"
)

type Amf struct {
	data    []byte
	amfType byte
	reader  *bytes.Reader
	res     []interface{}
	decoder Decoder
}

type Decoder interface {
	decode() (interface{}, error)
	isFinished() bool
}

func NewAmfObject(data []byte, amfType byte) *Amf {
	amf := &Amf{
		data:    data,
		amfType: amfType,
		reader:  bytes.NewReader(data),
	}
	var decoder Decoder
	if amfType == 0x00 {
		decoder = amf0Decoder{
			reader: amf.reader,
		}
	} else if amfType == 0x03 {
		decoder = amf0Decoder{ //we must change this to amf3
			reader: amf.reader,
		}
	}
	return amf
}

func (a *Amf) Decode() error {
	for {
		if a.decoder.isFinished() {
			break
		}
		res, err := a.decoder.decode()
		if err != nil {
			return err
		}
		a.res = append(a.res, res)
	}
	return nil
}
