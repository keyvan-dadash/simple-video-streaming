package chunk

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"

	"../../amf"

	"github.com/sirupsen/logrus"

	uuid "github.com/satori/go.uuid"
)

var (
	setChunkSizeID       = uint32(1)
	abortMessageID       = uint32(2)
	ackID                = uint32(3)
	userControlMessageID = uint32(4)
	windowAckSizeID      = uint32(5)
	setPeerBandWidth     = uint32(6)
	audioMessageID       = uint32(8)
	videoMessageID       = uint32(9)
	dataMessageAmf3ID    = uint32(15)
	sharedObjectAmf3ID   = uint32(16)
	amf3ID               = uint32(17)
	dataMessageAmf0ID    = uint32(18)
	sharedObjectAmf0ID   = uint32(19)
	amf0ID               = uint32(20)
	aggregateMessageID   = uint32(22)
)

//Conn is Custom Conn to handel connection
type Conn struct {
	net.Conn
	id                  string
	chunkSize           uint32
	remoteChunkSize     uint32
	windowAckSize       uint32
	remoteWindowAckSize uint32
	received            uint32
	ackReceived         uint32
	rw                  *ReaderWriter
	chunks              map[uint32]*Chunk

	bytesw *bytes.Buffer

	transactionID  int
	App            string `amf:"app" json:"app"`
	Flashver       string `amf:"flashVer" json:"flashVer"`
	SwfUrl         string `amf:"swfUrl" json:"swfUrl"`
	TcUrl          string `amf:"tcUrl" json:"tcUrl"`
	Fpad           bool   `amf:"fpad" json:"fpad"`
	AudioCodecs    int    `amf:"audioCodecs" json:"audioCodecs"`
	VideoCodecs    int    `amf:"videoCodecs" json:"videoCodecs"`
	VideoFunction  int    `amf:"videoFunction" json:"videoFunction"`
	PageUrl        string `amf:"pageUrl" json:"pageUrl"`
	ObjectEncoding int    `amf:"objectEncoding" json:"objectEncoding"`
}

//NewConn create Conn struct based on buffersize and conn that passed
func NewConn(c net.Conn, bufferSize int) *Conn {
	id, _ := uuid.NewV4()
	return &Conn{
		id:                  id.String(),
		Conn:                c,
		chunkSize:           128,
		remoteChunkSize:     128,
		windowAckSize:       2500000,
		remoteWindowAckSize: 2500000,
		rw:                  NewReaderWriter(bufio.NewReadWriter(bufio.NewReader(c), bufio.NewWriter(c))),
		chunks:              make(map[uint32]*Chunk),
		bytesw:              bytes.NewBuffer(nil),
	}
}

func (c *Conn) WriteChunk(chunk *Chunk) {

	numberOfChunks := (chunk.messageLength / c.chunkSize)

	logrus.Debug("number of blocks ", numberOfChunks)
	logrus.Debug(chunk.messageLength)
	logrus.Debug(c.chunkSize)

	sentSize := uint32(0)
	for i := uint32(0); i <= numberOfChunks; i++ {

		if sentSize == chunk.messageLength {
			break
		}
		if i == uint32(0) {
			chunk.fmt = 0
		} else {
			chunk.fmt = 3
		}
		chunk.writeHeader(c.rw)

		startPtr := c.chunkSize * i
		endPtr := startPtr + c.chunkSize
		if endPtr > uint32(len(chunk.data)) {
			endPtr = uint32(len(chunk.data))
		}

		buffer := chunk.data[startPtr:endPtr]

		c.rw.Write(buffer)

		sentSize += (endPtr - startPtr)
	}

}

func (c *Conn) ReadChunk(chunk *Chunk) {
	switch chunk.CSID {
	case 0:
		chunk.CSID = c.rw.ReadNByte(1) + 64
	case 1:
		chunk.CSID += c.rw.ReadNByte(2) + 64
	}
	logrus.Debugf("Got Chunk CSID: %v", chunk.CSID)

	switch chunk.fmt {
	case 0:
		chunk.timeStamp = c.rw.ReadNByte(3)
		logrus.Debugf("[Type0] timestamp: %v", chunk.timeStamp)
		chunk.messageLength = c.rw.ReadNByte(3)
		logrus.Debugf("[Type0] messagelength: %v", chunk.messageLength)
		chunk.messageTypeID = c.rw.ReadNByte(1)
		logrus.Debugf("[Type0] messagetypeid: %v", chunk.messageTypeID)
		chunk.messageStreamID = c.rw.ReadNByte(4)
		logrus.Debugf("[Type0] messagesteamid: %v", chunk.messageStreamID)
		if chunk.timeStamp == 0xffffff {
			chunk.haveExtendedTimeStamp = true
			chunk.timeStamp = c.rw.ReadNByte(4)
		} else {
			chunk.haveExtendedTimeStamp = false
		}

		chunk.data = make([]byte, chunk.messageLength)
		chunk.currentPos = 0
		//c.chunkSize = chunk.messageLength
		chunk.isFinished = false
		logrus.Debug("finished chunk type 0")
	case 1:
		chunk.timeStampDelta = c.rw.ReadNByte(3)
		chunk.messageLength = c.rw.ReadNByte(3)
		chunk.messageTypeID = c.rw.ReadNByte(1)

		if chunk.timeStampDelta == 0xffffff {
			chunk.haveExtendedTimeStamp = true
			chunk.timeStampDelta = c.rw.ReadNByte(4)
		} else {
			chunk.haveExtendedTimeStamp = false
		}

		chunk.timeStamp += chunk.timeStampDelta
		logrus.Debug("finished chunk type 3")
	case 2:
		chunk.timeStampDelta = c.rw.ReadNByte(3)

		if chunk.timeStamp == 0xffffff {
			chunk.haveExtendedTimeStamp = true
			chunk.timeStampDelta = c.rw.ReadNByte(4)
		} else {
			chunk.haveExtendedTimeStamp = false
		}

		chunk.timeStamp += chunk.timeStampDelta
		logrus.Debug("finished chunk type 2")
	case 3:
		//
	}

	logrus.Debugf("Start create buffer and copy chunk len : %v, current pos: %v, chunk type id : %v, chunk fmt : %v, chunk message type : %v",
		chunk.messageLength, chunk.currentPos, chunk.CSID, chunk.fmt, chunk.messageTypeID)

	size := uint32(0)
	if chunk.currentPos+c.remoteChunkSize > uint32(len(chunk.data)) {
		size = uint32(len(chunk.data))
	} else {
		size = chunk.currentPos + c.remoteChunkSize
	}
	buffer := chunk.data[chunk.currentPos:size]

	logrus.Debug("start reading buffer")
	if _, err := c.rw.Read(buffer); err != nil {
		logrus.Errorf("[Error] error during reading and copy to buffer err: %v", err)
	}
	logrus.Debug(buffer)
	logrus.Debugf("finished reading buffer %v", len(buffer))

	chunk.currentPos += uint32(len(buffer))
	if chunk.currentPos == chunk.messageLength {
		chunk.isFinished = true
	}

}

func (c *Conn) handleControlMessages(chunk *Chunk) {

	if chunk.messageTypeID == setChunkSizeID {
		fmt.Printf("chunk size is %v\n", binary.BigEndian.Uint32(chunk.data))
		c.remoteChunkSize = binary.BigEndian.Uint32(chunk.data)
	} else if chunk.messageTypeID == windowAckSizeID {
		c.remoteWindowAckSize = binary.BigEndian.Uint32(chunk.data)
	}
	//can add extra message control here
}

func (c *Conn) Ack(chunk *Chunk) {
	c.ackReceived += chunk.messageLength
	if c.ackReceived >= c.remoteWindowAckSize {
		ackChunk := CreateAck(*chunk)
		c.WriteChunk(ackChunk)
		c.ackReceived = 0
		logrus.Debug("ack sent successfully")
	}
}

func (c *Conn) HandleCmdMsg(chunk *Chunk) {
	amfType := amf.AMF0
	if chunk.messageTypeID == 17 {
		chunk.data = chunk.data[1:]
	}
	logrus.Debug(chunk.data)
	strlen := int(binary.BigEndian.Uint16(chunk.data[1:3]))
	s := string(chunk.data[3 : 3+strlen])
	logrus.Debug(s)
	r := bytes.NewReader(chunk.data)
	decoder := &amf.Decoder{}
	vs, err := decoder.DecodeBatch(r, amf.Version(amfType))
	if err != nil && err != io.EOF {
		logrus.Error(err)
	}
	//logrus.Debugf("rtmp req: %v", vs)
	logrus.Debugf("cmd msg %v", vs)
	switch vs[0].(type) {
	case string:
		switch vs[0].(string) {
		case "connect":
			c.connect(vs[1:])
			c.connectResp(chunk)
		default:
			logrus.Debug("no support command=", vs[0].(string))
		}
	}

}

func (c *Conn) connect(vs []interface{}) {
	for _, v := range vs {
		switch v.(type) {
		case string:
		case float64:
			id := int(v.(float64))
			if id != 1 {
				return
			}
			c.transactionID = id
			logrus.Debugf("transaction id is %v", id)
		case amf.Object:
			logrus.Debug("soooooo")
			obimap := v.(amf.Object)
			if app, ok := obimap["app"]; ok {
				c.App = app.(string)
			}
			if flashVer, ok := obimap["flashVer"]; ok {
				c.Flashver = flashVer.(string)
			}
			if tcurl, ok := obimap["tcUrl"]; ok {
				c.TcUrl = tcurl.(string)
			}
			if encoding, ok := obimap["objectEncoding"]; ok {
				c.ObjectEncoding = int(encoding.(float64))
			}
		}
	}
}

func CreateAck(chunk Chunk) *Chunk {

	ackChunk := &Chunk{
		fmt:             0,
		CSID:            2,
		messageTypeID:   ackID,
		messageStreamID: 0,
		messageLength:   chunk.messageLength,
		data:            make([]byte, 4),
	}

	binary.BigEndian.PutUint32(ackChunk.data, chunk.messageLength)

	return ackChunk

}

func (conn *Conn) NewSetChunkSize(size uint32) Chunk {
	return initControlMsg(setChunkSizeID, 4, size)
}

func (conn *Conn) NewWindowAckSize(size uint32) Chunk {
	return initControlMsg(windowAckSizeID, 4, size)
}

func (conn *Conn) NewSetPeerBandwidth(size uint32) Chunk {
	ret := initControlMsg(setPeerBandWidth, 5, size)
	ret.data[4] = 2
	return ret
}
func initControlMsg(id, size, value uint32) Chunk {
	ret := Chunk{
		fmt:             0,
		CSID:            2,
		messageTypeID:   id,
		messageStreamID: 0,
		messageLength:   size,
		data:            make([]byte, size),
	}
	binary.BigEndian.PutUint32(ret.data[:size], value)
	logrus.Debugf("init controll packet is %v and value", ret, value)
	return ret
}

func (conn *Conn) connectResp(chunk *Chunk) {
	c := conn.NewWindowAckSize(2500000)
	conn.WriteChunk(&c)
	c = conn.NewSetPeerBandwidth(2500000)
	conn.WriteChunk(&c)
	c = conn.NewSetChunkSize(uint32(1024))
	conn.chunkSize = 1024
	conn.WriteChunk(&c)
	conn.rw.Flush()

	resp := make(amf.Object)
	resp["fmsVer"] = "FMS/3,0,1,123"
	resp["capabilities"] = 31

	event := make(amf.Object)
	event["level"] = "status"
	event["code"] = "NetConnection.Connect.Success"
	event["description"] = "Connection succeeded."
	event["objectEncoding"] = conn.ObjectEncoding
	conn.writeMsg(chunk.CSID, chunk.messageStreamID, "_result", conn.transactionID, resp, event)
}

func (connServer *Conn) writeMsg(csid, streamID uint32, args ...interface{}) {
	connServer.bytesw.Reset()
	for _, v := range args {
		encoder := &amf.Encoder{}
		if _, err := encoder.Encode(connServer.bytesw, v, amf.AMF0); err != nil {
			logrus.Error(err)
			return
		}
	}
	msg := connServer.bytesw.Bytes()
	c := Chunk{
		fmt:             0,
		CSID:            csid,
		timeStamp:       0,
		messageTypeID:   20,
		messageStreamID: streamID,
		messageLength:   uint32(len(msg)),
		data:            msg,
	}
	logrus.Debug(c)
	connServer.WriteChunk(&c)
	connServer.rw.Flush()
}
