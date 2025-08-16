package main

import (
	"fmt"
	"log"
	"net"
	"os"
)

const CRLF = "\r\n"
const PORT = 4221

func main() {
	listen, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%v", PORT))
	if err != nil {
		fmt.Printf("Failed to bind to port %v\n", PORT)
		log.Fatal()
	}
	fmt.Printf("running on port %v\n", PORT)
	for {
		conn, err := listen.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		// creates a goroutine for concurrent connections
		go handleConnection(conn)
	}
}