package conn

import (
	"bufio"
	"bytes"
	"net"

	"../amf"

	uuid "github.com/satori/go.uuid"
)

var (
	DEFAULT_RTMP_BUFFER_SIZE     = uint32(128)
	DEFAULT_RTMP_WINDOW_ACK_SIZE = uint32(2500000)
)

//Conn is Conn that present Connection over network for rtmp
type Conn struct {
	net.Conn
	ID                  string
	ServerChunkSize     uint32
	ServerWindowAckSize uint32
	ClientChunkSize     uint32
	ClientWindowAckSize uint32
	AckReceived         uint32
	ReaderWriter        *bufio.ReadWriter
	Encoder             *amf.Encoder
	Decoder             *amf.Decoder
	amfVersion          amf.Version
	ConnInfo            ConnectInfo
	Bytesw              *bytes.Buffer
}

//NewConn create Conn struct based on buffersize and conn that passed
func NewConn(c net.Conn) *Conn {
	id, _ := uuid.NewV4()
	return &Conn{
		ID:                  id.String(),
		Conn:                c,
		ServerChunkSize:     DEFAULT_RTMP_BUFFER_SIZE,
		ClientChunkSize:     DEFAULT_RTMP_BUFFER_SIZE,
		ServerWindowAckSize: DEFAULT_RTMP_WINDOW_ACK_SIZE,
		ClientWindowAckSize: DEFAULT_RTMP_WINDOW_ACK_SIZE,
		ReaderWriter:        bufio.NewReadWriter(bufio.NewReader(c), bufio.NewWriter(c)),
		Encoder:             &amf.Encoder{},
		Decoder:             &amf.Decoder{},
		Bytesw:              bytes.NewBuffer(nil),
		AckReceived:         0,
	}
}

//SetAmfVersion set amf version
func (c *Conn) SetAmfVersion(ver amf.Version) {
	c.amfVersion = ver
}

//GetAmfVersion return amf version
func (c *Conn) GetAmfVersion() amf.Version {
	return c.amfVersion
}
