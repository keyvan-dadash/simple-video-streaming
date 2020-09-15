package rtmp

import (
	"net"

	"./chunk"

	"github.com/sirupsen/logrus"
)

//Server is struct that present server logics
type Server struct {
	listner net.Listener
}

//Start is function that start server and accept clients
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

		newConn := chunk.NewConn(conn, 1024)
		go s.handleConn(newConn)
	}
}

func (s *Server) handleConn(conn *chunk.Conn) {

	//Begin handshake process
	conn.HandShake()

	//Start to read Chunks
	chunk.Handler(conn)
}
