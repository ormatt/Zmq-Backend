package main

import "fmt"

import (
	zmq "github.com/pebbe/zmq4"
)

func brokerFunc(){
	logger := getNewLogger("Broker")

	frontend, _ := zmq.NewSocket(zmq.ROUTER)
	defer frontend.Close()
	frontend.Bind(fmt.Sprintf("%s://%s:%s", Proto, FrontendHost, FrontendPort))
	logger.Info("Frontend Binded")

	backend, _ := zmq.NewSocket(zmq.ROUTER)
	defer backend.Close()
	backend.Bind(fmt.Sprintf("%s://%s:%s", Proto, BackendHost, BackendPort))
	logger.Info("Backend binded")

	workerQueue := make([]string, 0, WorkerPoolSize)

	pollerBackend := zmq.NewPoller()
	pollerBackend.Add(backend, zmq.POLLIN)

	pollerBoth := zmq.NewPoller()
	pollerBoth.Add(backend, zmq.POLLIN)
	pollerBoth.Add(frontend, zmq.POLLIN)

	for {
		//  Poll frontend only if we have available workers
		var sockets []zmq.Polled
		if len(workerQueue) > 0 {
			sockets, _ = pollerBoth.Poll(-1)
		} else {
			sockets, _ = pollerBackend.Poll(-1)
		}

		for _, socket := range sockets {
			switch socket.Socket {
			//  Handle worker activity on backend
			case backend:
				//  Queue worker identity for load-balancing
				workerID, _ := backend.Recv(0)
				workerQueue = append(workerQueue, workerID)
				if empty, _ := backend.Recv(0); empty != ""{	
					reportError(fmt.Sprintf("Empty frame is not empty: %q", empty), logger)
				}

				//  Third frame is READY or else a client reply identity
				if clientID, _ := backend.Recv(0); clientID != "READY" {
					if empty, _ := backend.Recv(0); empty != "" {
						reportError(fmt.Sprintf("Empty frame is not empty: %q", empty), logger)
					}
					reply, _ := backend.Recv(0)
					frontend.Send(clientID, zmq.SNDMORE)
					frontend.Send("", zmq.SNDMORE)		//Send empty
					frontend.Send(reply, 0)
				}

			// handle a client req
			case frontend:
				//  Now get next client req, route to last-used worker
				//  Client req is [identity][empty][req]
				clientID, _ := frontend.Recv(0)
				if empty, _ := frontend.Recv(0); empty != "" {
					reportError(fmt.Sprintf("Empty frame is not empty: %q", empty), logger)
				}
				req, _ := frontend.Recv(0)
				logger.Info("RECV: ", req)


				backend.Send(workerQueue[0], zmq.SNDMORE)
				backend.Send("", zmq.SNDMORE)
				backend.Send(clientID, zmq.SNDMORE)
				backend.Send("", zmq.SNDMORE)
				backend.Send(req, 0)

				//  Dequeue and drop the next worker identity
				workerQueue = workerQueue[1:]

			}
		}
	}
}