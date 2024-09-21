package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
)

const (
	CONN_HOST = "localhost"
	CONN_PORT = "3333"
	CONN_TYPE = "tcp"
)

func main() {
	l, err := net.Listen(CONN_TYPE, CONN_HOST+":"+CONN_PORT)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	defer l.Close()
	fmt.Println("Listening on " + CONN_HOST + ":" + CONN_PORT)

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}
		go handleRequest(conn)
	}
}

func handleRequest(conn net.Conn) {
	buf, read_err := ioutil.ReadAll(conn)
	if read_err != nil {
		fmt.Println("failed:", read_err)
		return
	}
	fmt.Println("Got: ", string(buf))

	_, write_err := conn.Write([]byte("Message received.\n"))
	if write_err != nil {
		fmt.Println("failed:", write_err)
		return
	}
	conn.Close()
}
