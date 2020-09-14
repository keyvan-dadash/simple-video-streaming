package chunk

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"io"
	"time"

	"github.com/sirupsen/logrus"
)

var (
	timeout = 3 * time.Second
)

//HandShake is function that handshake in rtmp protocol
func (conn *Conn) HandShake() {

	//C0 and S0 packet in handshake
	C0 := make([]byte, 1)
	S0 := make([]byte, 1)

	//C1 packet in handshake
	C1 := make([]byte, 1536)

	//C2 packet in handshake
	C2 := make([]byte, 1536)

	//reading C0 and C1
	C0C1 := make([]byte, 1537)

	//look like server send C0 and C1 together
	conn.Conn.SetDeadline(time.Now().Add(timeout))
	if _, err := io.ReadFull(conn.rw, C0C1); err != nil {
		logrus.Errorf("[Error] %v error during reading C0 in handshake", err)
		return
	}
	logrus.Debug("[Debug] successfully read C0 and C1")

	//spreate C0 and C1
	C0[0] = C0C1[0]
	C1 = C0C1[1:]

	//Check for version 3
	if C0[0] != 0x3 {
		logrus.Errorf("[Error] cannot support verison %v", C0)
		return
	}

	//set verison
	S0[0] = 3

	//create basic of S1 packet
	S1Time := make([]byte, 4)
	S1Zero := make([]byte, 4)
	S1Rand := make([]byte, 1528)

	//inilize basic of S1 basics
	binary.BigEndian.PutUint32(S1Time, uint32(time.Now().Unix()))
	binary.BigEndian.PutUint32(S1Zero, uint32(0))
	rand.Read(S1Rand)

	//create S1
	S1 := []byte{}
	S1 = append(S1, S1Time...)
	S1 = append(S1, S1Zero...)
	S1 = append(S1, S1Rand...)

	if len(S1) != 1536 {
		logrus.Error("[Error] S1 MUST be 1536 size")
		return
	}

	//create basic of S2 packet
	S2Time := make([]byte, 4)
	S2Time2 := make([]byte, 4)
	S2Rand := make([]byte, 1528)

	//inilize basic of S2 basics
	copy(S2Time, C1[0:4])
	copy(S2Time2, C1[0:4])
	copy(S2Rand, C1[8:])

	//create S2
	S2 := []byte{}
	S2 = append(S2, S2Time...)
	S2 = append(S2, S2Time2...)
	S2 = append(S2, S2Rand...)

	//craete total handshake packet
	S0S1S2 := []byte{}
	S0S1S2 = append(S0S1S2, S0...)
	S0S1S2 = append(S0S1S2, S1...)
	S0S1S2 = append(S0S1S2, S2...)

	//write total handshake packet to connection
	conn.Conn.SetDeadline(time.Now().Add(timeout))
	if _, err := conn.rw.Write(S0S1S2); err != nil {
		logrus.Errorf("[Error] %v during sending S2 in handshake", err)
	}

	conn.rw.Flush()
	logrus.Debug("[Debug] successfully send total handshake packet")

	//reading C2
	conn.Conn.SetDeadline(time.Now().Add(timeout))
	if _, err := io.ReadFull(conn.rw, C2); err != nil {
		logrus.Errorf("[Error] %v error during reading C2 in handshake", err)
		return
	}

	logrus.Debug("[Debug[ successfully read C2 packet")

	//validate C2 packet
	if bytes.Compare(C2[8:], S1Rand) != 0 {
		logrus.Error("[Error] C2 packet random field validation failed")
		return
	}

	if bytes.Compare(C2[0:4], S1Time) != 0 {
		logrus.Error("[Error] C2 packet time field validation failed")
		return
	}

	logrus.Debugf("[Debug] handshake done with client %v", conn.id)
}
