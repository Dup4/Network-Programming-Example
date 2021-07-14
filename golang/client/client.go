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
	"time"

	"github.com/Dup4/Socket-Network-Programming-Learning/common"
)

const unixSocketDomain = "unix"

func main() {
	input := bufio.NewScanner(os.Stdin)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	closing := make(chan struct{})
	closeWg := new(sync.WaitGroup)

	closeWg.Add(1)
	go func() {
		conn, reader := getConn()

		go func() {
			for {
				fmt.Print("Please enter the data to be sent: ")

				input.Scan()
				data := input.Text()

				conn.Write([]byte(data + "\n"))

				msg, err := reader.ReadString('\n')

				if err != nil {
					log.Println(err)
					conn.Close()
					conn, reader = getConn()
				} else {
					fmt.Println(msg)
				}
			}
		}()

		<-closing
		if conn != nil {
			conn.Close()
			closeWg.Done()
		}
	}()

	<-sigCh
	close(closing)
	closeWg.Wait()

	os.Exit(0)
}

func getConn() (net.Conn, *bufio.Reader) {
	var conn net.Conn

	// Retry on failure
	for {
		var err error
		conn, err = initConn()

		if err != nil {
			log.Println(err)
			time.Sleep(time.Second * 1)
		} else {
			break
		}
	}

	return conn, bufio.NewReader(conn)
}

func initConn() (net.Conn, error) {
	conn, err := net.Dial(unixSocketDomain, common.SocketPath)

	if err != nil {
		return nil, err
	}

	return conn, nil
}
