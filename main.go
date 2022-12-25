package main

import (
	"fmt"
	"os"
	"os/signal"
	"time"
	"strconv"

	"github.com/dachunky/echoserver/pkg/echoserver"
	"github.com/dachunky/echoserver/pkg/logging"
)

func main() {
	stopSignal := make(chan os.Signal, 1)
	signal.Notify(stopSignal, os.Interrupt)
	exit := make(chan bool)
	go func() {
		<-stopSignal
		logging.Log(logging.LOG_MAIN, "CTRL+C retrieved")
		exit <- true
	}()
	// read cmd line paramter
	args := os.Args
	port := 12345
	if len(args) > 1 {
		fmt.Printf("retrieve %d args\n", len(args))
    	_, err := strconv.Atoi(args[1])
		if err == nil {
			port, _ = strconv.Atoi(args[1])
		} else {
			fmt.Printf("failed to parse port: %s\n", args[1])
			logging.LogFmt(logging.LOG_ERROR, "invalid port provided: %s", args[1])
	   	}
	}
	fmt.Printf("Hello There - listening on port: %d\n", port)
	es := echoserver.NewEchoServer(port)
	err := es.StartListening()
	if err != nil {
		logging.Log(logging.LOG_FATAL, "server start failed")
		fmt.Println("failed to start server")
		return
	}
	for {
		select {
		case <-exit:
			fmt.Println("exit signal retrieved. Stop server ...")
			es.Stop()
			fmt.Println("... server stopped")
			return
		default:
			err = es.IsHealthy()
			if err != nil {
				logging.Log(logging.LOG_FATAL, "retrieve unhealthy state of the server")
				fmt.Println("server unhealthy -> quit")
				return
			}
			time.Sleep(1000)
		}
	}
}
