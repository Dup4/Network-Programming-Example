package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"network-example/common"
)

var closeWg sync.WaitGroup
var closing chan struct{}

func main() {
	uds := initUds()
	defer uds.Close()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	closing = make(chan struct{})

	go func() {
		for {
			conn, err := uds.Accept()

			if err != nil {
				log.Fatal("Request acceptance error ->", err)
			}

			closeWg.Add(1)
			log.Println("A connection is established.")
			go handle(conn)
		}
	}()

	<-sigCh
	close(closing)
	closeWg.Wait()

	os.Exit(0)
}

func initUds() net.Listener {
	uds, err := net.Listen("unix", common.SocketPath)

	if err != nil {
		log.Println("Unix Domain Socket creation failed，attempting to recreate ->", err)
		err = os.Remove(common.SocketPath)

		if err != nil {
			log.Fatalln("Failed to delete socket file, the program exits ->", err)
		}

		return initUds()
	} else {
		log.Println("Successfully created Unix Domain Socket.")
	}

	return uds
}

func handle(conn net.Conn) {
	defer func() {
		conn.Close()
		log.Println("A connection is closed.")
		closeWg.Done()
	}()

	reader := bufio.NewReader(conn)

	go func() {
		for {
			msg, err := reader.ReadString('\n')

			if err != nil {
				// If something goes wrong, disconnect the link.
				// Hand it over to the client to reconnect.
				return
			}

			fmt.Println("Receive a message from the remote end: ", msg)
			conn.Write([]byte(msg))
		}
	}()

	<-closing
}
