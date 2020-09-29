package main

import (
	"./protocols/rtmp/server"
	"github.com/sirupsen/logrus"
)

func main() {

	logrus.SetLevel(logrus.DebugLevel)

	c := make(chan int)

	s := server.NewRtmpServer("127.0.0.1", 1935)

	s.StartServer()

	<-c
}
