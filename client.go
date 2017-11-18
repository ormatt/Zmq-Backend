package main

//go:generate msgp

import (
	"time"
	"strconv"
	//"github.com/tinylib/msgp/msgp"
	zmq "github.com/pebbe/zmq4"
	//log "github.com/sirupsen/logrus"
	"fmt"
)


//type genericMsg struct {
//	Uuid string  `msg:"uuid"`
//	Data string  `msg:"data"`
//	TS float64 `msg:"timestamp"`
//}

var (
	timeoutSeconds = 1000 * time.Millisecond
	tickSeconds = 500 * time.Millisecond
)


func clientFunc(finishChan chan bool, clientNum int) {
	logger := getNewLogger("Client")
	timeout := time.After(timeoutSeconds)
	tick := time.Tick(tickSeconds)

	client, _ := zmq.NewSocket(zmq.REQ)
	defer client.Close()
	client.Connect(fmt.Sprintf("%s://%s:%s", Proto, FrontendHost, FrontendPort))

	for {
		select {
		case <-timeout:
			finishChan <- true
			//logger.Info("Timeout reached!")
			return
		case <-tick:
			req := strconv.Itoa(clientNum) + "hello"
			client.Send(req, 0)

			//fmt.Println("[client_task] SEND: ", msg)
			resp, _ := client.Recv(0)
			logger.Info("RECV: ", resp)
		}
	}
}
