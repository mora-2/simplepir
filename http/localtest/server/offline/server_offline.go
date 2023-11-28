// server_offline.go
package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"

	"github.com/mora-2/simplepir/http/localtest/server/config"
)

func handleConnection(conn net.Conn, shared_data config.Shared_data) {
	defer conn.Close()

	encoder := json.NewEncoder(conn)
	err := encoder.Encode(shared_data)
	if err != nil {
		fmt.Println("Error encoding shared_data:", err.Error())
	}
	fmt.Println("Offline shared_data sent.")

}

func main() {
	// load file
	shared_file, err := os.Open("../data/shared_data")
	if err != nil {
		fmt.Println("Error opening shared_data_file:", err.Error())
	}
	defer shared_file.Close()

	decoder := json.NewDecoder(shared_file)
	var shared_data config.Shared_data
	err = decoder.Decode(&shared_data)
	if err != nil {
		fmt.Println("Error loading shared_data:", err.Error())
	}

	// start listening
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		return
	}
	defer listener.Close()
	fmt.Println("Server listening on :8080")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err.Error())
			break
		}

		go handleConnection(conn, shared_data)
	}
}
