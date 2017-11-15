package main

import "fmt"

import (
	zmq "github.com/pebbe/zmq4"
)

func brokerFunc(){
	logger := getNewLogger("Broker")

	frontend, _ := zmq.NewSocket(zmq.ROUTER)
	defer frontend.Close()
	frontend.Bind(fmt.Sprintf("%s://%s:%s", proto, frontendHost, frontendPort))
	logger.Info("[fronted] binded")

	backend, _ := zmq.NewSocket(zmq.ROUTER)
	defer backend.Close()
	backend.Bind(fmt.Sprintf("%s://%s:%s", proto, backendHost, backendPort))
	logger.Info("[backend] binded")

	workerQueue := make([]string, 0, workerPoolSize)

	poller1 := zmq.NewPoller()
	poller1.Add(backend, zmq.POLLIN)

	poller2 := zmq.NewPoller()
	poller2.Add(backend, zmq.POLLIN)
	poller2.Add(frontend, zmq.POLLIN)

	for {
		//  Poll frontend only if we have available workers
		var sockets []zmq.Polled
		if len(workerQueue) > 0 {
			sockets, _ = poller2.Poll(-1)
			//logger.Info("1 Worker queue=", len(workerQueue))
		} else {
			sockets, _ = poller1.Poll(-1)
			//logger.Info("2 Worker queue=", len(workerQueue))
		}

		for _, socket := range sockets {
			switch socket.Socket {
			//  Handle worker activity on backend
			case backend:
				//  Queue worker identity for load-balancing
				workerID, _ := backend.Recv(0)
				if !(len(workerQueue) < workerPoolSize) {
					panic("!(len(workerQueue) < NBR_WORKERS)")
				}
				workerQueue = append(workerQueue, workerID)

				backend.Recv(0)	//  Second frame is empty
				//  Third frame is READY or else a client reply identity
				clientID, _ := backend.Recv(0)

				//  If client reply, send rest back to frontend
				if clientID != "READY" {
					empty, _ := backend.Recv(0)
					if empty != "" {
						panic(fmt.Sprintf("empty is not \"\": %q", empty))
					}
					reply, _ := backend.Recv(0)
					frontend.Send(clientID, zmq.SNDMORE)
					frontend.Send("", zmq.SNDMORE)		//Send empty
					frontend.Send(reply, 0)
				}

			// handle a client request
			case frontend:
				//  Now get next client request, route to last-used worker
				//  Client request is [identity][empty][request]
				clientID, _ := frontend.Recv(0)
				frontend.Recv(0)	//Recv empty
				request, _ := frontend.Recv(0)

				backend.Send(workerQueue[0], zmq.SNDMORE)
				backend.Send("", zmq.SNDMORE)
				backend.Send(clientID, zmq.SNDMORE)
				backend.Send("", zmq.SNDMORE)
				backend.Send(request, 0)

				//  Dequeue and drop the next worker identity
				workerQueue = workerQueue[1:]

			}
		}
	}
}