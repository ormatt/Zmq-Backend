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

	//for client_nbr > 0 {
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
			case backend:
				//  Handle worker activity on backend
				//  Queue worker identity for load-balancing
				worker_id, _ := backend.Recv(0)
				//logger.Info("[backend] Recv worker_id: ", worker_id)
				if !(len(workerQueue) < workerPoolSize) {
					panic("!(len(workerQueue) < NBR_WORKERS)")
				}
				workerQueue = append(workerQueue, worker_id)

				//  Second frame is empty
				empty, _ := backend.Recv(0)
				if empty != "" {
					panic(fmt.Sprintf("empty is not \"\": %q", empty))
				}

				//  Third frame is READY or else a client reply identity
				client_id, _ := backend.Recv(0)
				//logger.Info("[backend] Recv client_id: ", client_id)

				//  If client reply, send rest back to frontend
				if client_id != "READY" {
					empty, _ := backend.Recv(0)
					if empty != "" {
						panic(fmt.Sprintf("empty is not \"\": %q", empty))
					}
					reply, _ := backend.Recv(0)
					frontend.Send(client_id, zmq.SNDMORE)
					frontend.Send("", zmq.SNDMORE)		//Send empty
					frontend.Send(reply, 0)
				}

			case frontend:
				//logger.Info("b")

				//  Here is how we handle a client request:

				//  Now get next client request, route to last-used worker
				//  Client request is [identity][empty][request]
				client_id, _ := frontend.Recv(0)
				//logger.Info("[frontend] Recv client_id: ", client_id)
				empty, _ := frontend.Recv(0)
				//logger.Info("[frontend] Recv empty: ", empty)
				if empty != "" {
					panic(fmt.Sprintf("empty is not \"\": %q", empty))
				}
				request, _ := frontend.Recv(0)

				backend.Send(workerQueue[0], zmq.SNDMORE)
				//logger.Info("[frontend] Send workerQueue: my...")
				backend.Send("", zmq.SNDMORE)
				//logger.Info("[frontend] Send empty: ")
				backend.Send(client_id, zmq.SNDMORE)
				//logger.Info("[frontend] Send client_id: ", client_id)
				backend.Send("", zmq.SNDMORE)
				//logger.Info("[frontend] Send empty: ")
				backend.Send(request, 0)
				//logger.Info("[frontend] Send request: ", request)

				//  Dequeue and drop the next worker identity
				workerQueue = workerQueue[1:]

			}
		}
	}
}