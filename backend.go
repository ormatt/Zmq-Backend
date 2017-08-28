package main

import (
	zmq "github.com/pebbe/zmq4"
	"fmt"
	"math/rand"
	"time"
	"strconv"
	"log"
)

const (
	NBR_CLIENTS = 1
	NBR_WORKERS = 3
)

var (
	r = rand.New(rand.NewSource(4))
)

func handleError(err error){
	if err != nil {
		log.Fatal(err)
	}
}

//  Basic request-reply client using REQ socket
//  Since Go Send and Recv can handle 0MQ binary identities we
//  don't need printable text identity to allow routing.

func client_task(finishChan chan<- bool) {
	client, _ := zmq.NewSocket(zmq.REQ)
	defer client.Close()
	client.Connect("tcp://127.0.0.1:5580")
	time.Sleep(3 * time.Second)
	timeout := time.After(10 * time.Second)
	tick := time.Tick(500 * time.Millisecond)

	for {
		select {
		case <-timeout:
			finishChan <- true
			fmt.Println("[client_task] Timeout!")
			return
		case <-tick:
			// set_id(client) //  Set a printable identity
			//fmt.Println("[client_task] Connected")

			//  Send request, get reply

			//msg := "HELLO"
			randMsg := strconv.Itoa(r.Intn(10))
			client.Send(randMsg, 0)
			fmt.Println("[client_task] SEND: ", randMsg)
			reply, _ := client.Recv(0)
			fmt.Println("[client_task] R E C V: ", reply)
		}
	}
}

//  While this example runs in a single process, that is just to make
//  it easier to start and stop the example.
//  This is the worker task, using a REQ socket to do load-balancing.
//  Since Go Send and Recv can handle 0MQ binary identities we
//  don't need printable text identity to allow routing.

func worker_task() {
	worker, _ := zmq.NewSocket(zmq.REQ)
	defer worker.Close()
	// set_id(worker)
	worker.Connect("tcp://127.0.0.1:5581")
	//fmt.Println("[worker_task] Connected")

	//  Tell broker we're ready for work
	//msg := "READY"
	//worker.Send(msg, 0)

	//fmt.Println("[worker_task] Send: ", msg)


	for {
		//  Read and save all frames until we get an empty frame
		//  In this example there is only 1 but it could be more
		identity, _ := worker.Recv(0)
		//fmt.Println("[worker_task] Recv identity: ", identity)
		empty, _ := worker.Recv(0)
		//fmt.Println("[worker_task] Recv empty: ", empty)
		if empty != "" {
			panic(fmt.Sprintf("empty is not \"\": %q", empty))
		}

		//  Get request, send reply
		request, _ := worker.Recv(0)
		//fmt.Println("\t[worker_task] R E C V: ", request, " ", len(request))

		worker.Send(identity, zmq.SNDMORE)
		//fmt.Println("[worker_task] Send identity: ", identity)
		worker.Send("", zmq.SNDMORE)
		//fmt.Println("[worker_task] Send empty: ")

		//OK := "OK"
		if len(request)>0 {
			randMsg, err := strconv.Atoi(request)
			handleError(err)
			randMsgReply := randMsg * 2
			worker.Send(fmt.Sprint(randMsg, " ", randMsgReply), 0)
			fmt.Println("\t[worker_task] SEND: ", randMsg, " ", randMsgReply)
		}else {
			worker.Send("", 0)
		}
	}
}

//  This is the main task. It starts the clients and workers, and then
//  routes requests between the two layers. Workers signal READY when
//  they start; after that we treat them as ready when they reply with
//  a response back to a client. The load-balancing data structure is
//  just a queue of next available workers.

func main() {
	//  Prepare our sockets
	frontend, _ := zmq.NewSocket(zmq.ROUTER)
	backend, _ := zmq.NewSocket(zmq.ROUTER)
	defer frontend.Close()
	defer backend.Close()
	frontend.Bind("tcp://127.0.0.1:5580")
	fmt.Println("[fronted] binded")
	backend.Bind("tcp://127.0.0.1:5581")
	fmt.Println("[backend] binded")
	finishChan := make(chan bool, 1)
	//wg.Add(1)

	client_nbr := 0
	for ; client_nbr < NBR_CLIENTS; client_nbr++ {
		go client_task(finishChan)
	}
	for worker_nbr := 0; worker_nbr < NBR_WORKERS; worker_nbr++ {
		go worker_task()
	}

	//  Here is the main loop for the least-recently-used queue. It has two
	//  sockets; a frontend for clients and a backend for workers. It polls
	//  the backend in all cases, and polls the frontend only when there are
	//  one or more workers ready. This is a neat way to use 0MQ's own queues
	//  to hold messages we're not ready to process yet. When we get a client
	//  reply, we pop the next available worker, and send the request to it,
	//  including the originating client identity. When a worker replies, we
	//  re-queue that worker, and we forward the reply to the original client,
	//  using the reply envelope.

	//  Queue of available workers
	worker_queue := make([]string, 0, 10)

	poller1 := zmq.NewPoller()
	poller1.Add(backend, zmq.POLLIN)
	poller2 := zmq.NewPoller()
	poller2.Add(backend, zmq.POLLIN)
	poller2.Add(frontend, zmq.POLLIN)

	//for client_nbr > 0 {
	for {
		select {
		case <-finishChan:
			fmt.Println("[main] Timeout!")
			break
		default:
			//  Poll frontend only if we have available workers
			var sockets []zmq.Polled
			if len(worker_queue) > 0 {
				sockets, _ = poller2.Poll(-1)
				//fmt.Println("1 Worker queue=", len(worker_queue))
			} else {
				sockets, _ = poller1.Poll(-1)
				//fmt.Println("2 Worker queue=", len(worker_queue))
			}
			for _, socket := range sockets {
				select {
				case b := <-finishChan:
					fmt.Println("[main2] Timeout!", b)
					break
				default:
					break
				}

				//fmt.Println("a")
				switch socket.Socket {
				case backend:

					//  Handle worker activity on backend
					//  Queue worker identity for load-balancing
					worker_id, _ := backend.Recv(0)
					//fmt.Println("[backend] Recv worker_id: ", worker_id)
					if !(len(worker_queue) < NBR_WORKERS) {
						panic("!(len(worker_queue) < NBR_WORKERS)")
					}
					worker_queue = append(worker_queue, worker_id)

					//  Second frame is empty
					empty, _ := backend.Recv(0)
					//fmt.Println("[backend] Recv empty: ", empty)
					if empty != "" {
						panic(fmt.Sprintf("empty is not \"\": %q", empty))
					}

					//  Third frame is READY or else a client reply identity
					client_id, _ := backend.Recv(0)
					//fmt.Println("[backend] Recv client_id: ", client_id)

					//  If client reply, send rest back to frontend
					if client_id != "READY" {
						empty, _ := backend.Recv(0)
						//fmt.Println("[backend] Recv empty: ", empty)
						if empty != "" {
							panic(fmt.Sprintf("empty is not \"\": %q", empty))
						}
						reply, _ := backend.Recv(0)
						frontend.Send(client_id, zmq.SNDMORE)
						//fmt.Println("[backend] Send client_id: ", client_id)
						frontend.Send("", zmq.SNDMORE)
						//fmt.Println("[backend] Send empty: ", empty)
						frontend.Send(reply, 0)
						fmt.Println("[backend] Send reply: ", reply)
						//client_nbr--
					}

				case frontend:
					//fmt.Println("b")

					//  Here is how we handle a client request:

					//  Now get next client request, route to last-used worker
					//  Client request is [identity][empty][request]
					client_id, _ := frontend.Recv(0)
					//fmt.Println("[frontend] Recv client_id: ", client_id)
					empty, _ := frontend.Recv(0)
					//fmt.Println("[frontend] Recv empty: ", empty)
					if empty != "" {
						panic(fmt.Sprintf("empty is not \"\": %q", empty))
					}
					request, _ := frontend.Recv(0)

					backend.Send(worker_queue[0], zmq.SNDMORE)
					//fmt.Println("[frontend] Send worker_queue: my...")
					backend.Send("", zmq.SNDMORE)
					//fmt.Println("[frontend] Send empty: ")
					backend.Send(client_id, zmq.SNDMORE)
					//fmt.Println("[frontend] Send client_id: ", client_id)
					backend.Send("", zmq.SNDMORE)
					//fmt.Println("[frontend] Send empty: ")
					backend.Send(request, 0)
					fmt.Println("[frontend] Send request: ", request)

					//  Dequeue and drop the next worker identity
					worker_queue = worker_queue[1:]

				}
			}
		}
	}

	time.Sleep(500 * time.Millisecond)
}

/*
func set_id(soc *zmq.Socket) {
	identity := fmt.Sprintf("%04X-%04X", rand.Intn(0x10000), rand.Intn(0x10000))
	soc.SetIdentity(identity)
}
*/