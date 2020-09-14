package main

import (
	"net"

	"github.com/sirupsen/logrus"

	"./protocols/rtmp/chunk"
)

func main() {

	lan, _ := net.Listen("tcp", "127.0.0.1:1935")

	logrus.SetLevel(logrus.DebugLevel)

	for {

		conn, _ := lan.Accept()
		nConn := chunk.NewConn(conn, 1024)

		nConn.HandShake()
	}
}
