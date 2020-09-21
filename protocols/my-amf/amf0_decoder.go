package amf

import (
	"bytes"
)

var (
	numberMarker      = 0x00
	booleanMarker     = 0x01
	stringMarker      = 0x02
	objectMarker      = 0x03
	movieclipMarker   = 0x04
	nullMarker        = 0x05
	undefinedMarker   = 0x06
	refrenceMarker    = 0x07
	ecmaArrayMarker   = 0x08
	objectEndMarker   = 0x0A
	dateMarker        = 0x0B
	longStringMarker  = 0x0C
	unsupportedMarker = 0x0D
	recordsetMarker   = 0x0E
	xmlDocumentMarker = 0x0F
	typedObjectMarker = 0x10
)

type amf0Decoder struct {
	reader *bytes.Reader
}

func (a *amf0Decoder) decode() (interface{}, error) {

	mraker, err := a.reader.ReadByte()
	if err != nil {
		return nil, err
	}
	switch int(mraker) {
	case numberMarker:
	case booleanMarker:
	case stringMarker:
	case objectMarker:
	case movieclipMarker:
	case nullMarker:
	case undefinedMarker:
	case refrenceMarker:
	case ecmaArrayMarker:
	case objectEndMarker:
	case dateMarker:
	case longStringMarker:
	case unsupportedMarker:
	case recordsetMarker:
	case xmlDocumentMarker:
	case typedObjectMarker:
	}
}
