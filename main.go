package main

import (
	"./protocols/rtmp"

	"github.com/sirupsen/logrus"
)

func main() {

	logrus.SetLevel(logrus.DebugLevel)

	c := make(chan int)

	s := rtmp.Server{}

	s.Start("127.0.0.1:1935", c)

	<-c
}
