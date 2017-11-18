package main

import (
	"fmt"
	zmq "github.com/pebbe/zmq4"

)

func handleReq(req string) string {
	var resp string
	if len(req)>0 {
		resp  = fmt.Sprintf("%s echo", req)
	}else{
		resp = ""
	}
	return resp
}

func workerFunc(workerNumber int) {
	logger := getNewLogger(fmt.Sprintf("Worker%d", workerNumber))
	worker, _ := zmq.NewSocket(zmq.REQ)
	defer worker.Close()
	worker.Connect(fmt.Sprintf("%s://%s:%s", Proto, BackendHost, BackendPort))
	worker.Send("READY", 0)

	for {
		identity, _ := worker.Recv(0)
		if empty, _ := worker.Recv(0); empty != ""{	//  Second frame is empty
			reportError(fmt.Sprintf("Empty frame is not empty: %q", empty), logger)
		}
		req, _ := worker.Recv(0)
		logger.Info("RECV: ", req)

		resp := handleReq(req)

		worker.Send(identity, zmq.SNDMORE)
		worker.Send("", zmq.SNDMORE)	// Send an empty frame
		worker.Send(resp, 0)
		//logger.Info("SEND: ", resp)
	}
}
