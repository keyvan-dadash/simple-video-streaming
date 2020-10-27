package conn

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"io"
	"time"

	"github.com/sirupsen/logrus"
)

//HandShakeConn handle handshake between client and server
type HandShakeConn struct {
	conn *Conn
}

//NewHandShakeConn return HandShakeConn for handling handshake in conn
func NewHandShakeConn(conn *Conn) *HandShakeConn {
	return &HandShakeConn{
		conn: conn,
	}
}

//HandShake the connection
func (hs *HandShakeConn) HandShake() {

	//client handshake packet
	C0 := make([]byte, 1)
	C1 := make([]byte, 1536)
	C2 := make([]byte, 1536)

	//server handshake packet
	S0 := make([]byte, 1)
	S1 := make([]byte, 1536)
	S2 := make([]byte, 1536)

	//read C0 and C1 together
	hs.readPacket(C0)
	logrus.Debug("finished reading packet C0 and validation is: %v", validateC0Packet(C0))

	hs.readPacket(C1)
	logrus.Debug("finished reading packet C1")

	//fill S0 and S1 and S2 packet
	fillS0Packet(S0)
	fillS1Packet(S1)
	fillS2Packet(S2, C1)
	logrus.Debug("finished filling packets")

	//write S0 and S1 and S2 packet
	hs.writePacket(S0)
	logrus.Debug("successfully sent S0 packet")
	hs.writePacket(S1)
	logrus.Debug("successfully sent S1 packet")
	hs.writePacket(S2)
	logrus.Debug("successfully sent S2 packet")

	//read C2 packet and begin validation
	hs.readPacket(C2)
	logrus.Debug("finished reading C2 packet")
	if !validateC2Packet(C2, S1) {
		logrus.Debug("validation of C2 packet fail start to close connection")
		hs.conn.Close()
		return
	}
	logrus.Debug("Handshake with client successfuly")

}

func (hs *HandShakeConn) readPacket(inBuffer []byte) {
	if _, err := io.ReadFull(hs.conn.ReaderWriter, inBuffer); err != nil {
		logrus.Errorf("[Error] %v error during reading buffer with size %v in handshake", err, len(inBuffer))
		return
	}
}

func (hs *HandShakeConn) writePacket(outBuffer []byte) {
	if _, err := hs.conn.ReaderWriter.Write(outBuffer); err != nil {
		logrus.Errorf("[Error] %v during sending buffer with size %v in handshake", err, len(outBuffer))
		return
	}
	hs.conn.ReaderWriter.Flush()
}

func validateC0Packet(C0 []byte) bool {
	if C0[0] != 0x3 {
		logrus.Errorf("[Error] cannot support verison %v", C0)
		return false
	}
	return true
}

func validateC2Packet(C2 []byte, S1 []byte) bool {
	if bytes.Compare(C2[8:], S1[8:]) != 0 {
		logrus.Error("[Error] C2 packet random field validation failed")
		return false
	}

	if bytes.Compare(C2[0:4], S1[0:4]) != 0 {
		logrus.Error("[Error] C2 packet time field validation failed")
		return false
	}

	return true
}

func fillS0Packet(S0 []byte) {
	S0[0] = 0x3
}

func fillS1Packet(S1 []byte) {
	binary.BigEndian.PutUint32(S1[0:4], uint32(time.Now().Unix()))
	binary.BigEndian.PutUint32(S1[4:8], uint32(0))
	rand.Read(S1[8:])
}

func fillS2Packet(S2 []byte, C1 []byte) {
	copy(S2[0:4], C1[0:4])
	copy(S2[4:8], C1[0:4])
	copy(S2[8:], C1[8:])
}
