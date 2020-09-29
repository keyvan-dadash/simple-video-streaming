package server

import (
	"net"
	"strconv"

	"github.com/sirupsen/logrus"
)

//RtmpServer is structure that present behaviour of rtmp server
type RtmpServer struct {
	Addr    string
	Port    int
	listner net.Listener
	Conns   map[string]Conn
}

//NewRtmpServer return RtmpServer with given addr and port
func NewRtmpServer(addr string, port int) *RtmpServer {
	return &RtmpServer{
		Addr:    addr,
		Port:    port,
		listner: nil,
	}
}

//StartServer start listen on addr with given port then in new groutine
func (rtmpS *RtmpServer) StartServer() {
	defer func() {
		rtmpS.listner.Close()
	}()
	if rtmpS.listner == nil {
		listner, err := net.Listen("tcp4", rtmpS.Addr+":"+strconv.Itoa(rtmpS.Port))
		if err != nil {
			logrus.Errorf("[Error] cannot start server err: %v", err)
			return
		}
		rtmpS.listner = listner
	}
	for {
		conn, err := rtmpS.listner.Accept()
		if err != nil {
			logrus.Errorf("[Error] cannot accept connectin from addr %v on port %v err: %v", rtmpS.Addr, rtmpS.Port, err)
			continue
		}
		connHandler := NewConnHadler(NewConn(conn))
		go connHandler.Handle()
	}
}
