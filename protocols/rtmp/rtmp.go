package rtmp

import (
	"net"

	"github.com/sirupsen/logrus"
)

//Server is struct that present server logics
type Server struct {
	listner net.Listener
}

//Start is function that start rtmp server and accept clients
func (s *Server) Start(addr string, end chan int) {
	defer func() {
		end <- 1
	}()

	lst, err := net.Listen("tcp", addr)

	if err != nil {
		logrus.Fatalf("[Fatal] cannot start server on %v because of %v", addr, err)
		return
	}
	s.listner = lst

	for {
		conn, err := s.listner.Accept()

		if err != nil {
			logrus.Errorf("[Error] client cannot connect err: %v", err)
		}

		go s.HandleConn(conn)
	}
}

//HandleConn handle single connection
func (s *Server) HandleConn(conn net.Conn) {
}
