package main

import (
	//"time"
	"log"
)

const (
	clientPoolSize = 10
	workerPoolSize = 5
	proto          = "tcp"
	frontendHost   = "127.0.0.1"
	frontendPort   = "5580"
	backendHost    = "127.0.0.1"
	backendPort    = "5581"

)

func handleError(err error){
	if err != nil {
		log.Fatal(err)
	}
}


func main() {
	logger := getNewLogger("Manager")
	finishChan := make(chan bool, clientPoolSize)

	for clientID := 0; clientID < clientPoolSize; clientID++ {
		go clientFunc(finishChan, clientID)
	}
	for workerID := 0; workerID < workerPoolSize; workerID++ {
		go workerFunc(workerID)
	}

	go brokerFunc()

	clientsLeft := clientPoolSize
	for {
		select {
		case <-finishChan:
			clientsLeft --
		default:
			if clientsLeft == 0 {
				logger.Info("[main] No clients left!")
				return
			} else {
				break
			}
		}
	}

}
