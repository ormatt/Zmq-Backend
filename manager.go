package main


func main() {
	logger := getNewLogger("Manager")
	finishChan := make(chan bool, ClientPoolSize)

	for clientNum := 0; clientNum < ClientPoolSize; clientNum++ {
		go clientFunc(finishChan, clientNum)
	}
	for workerNum := 0; workerNum < WorkerPoolSize; workerNum++ {
		go workerFunc(workerNum)
	}

	go brokerFunc()

	clientsLeft := ClientPoolSize
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
