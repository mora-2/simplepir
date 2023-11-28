// server.go
package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"

	// "runtime"
	// "time"

	"github.com/mora-2/simplepir/http/localtest/server/config"
	"github.com/mora-2/simplepir/pir"
)

func handleConnection(conn net.Conn, DB *pir.Database, server_state pir.State, shared_state pir.State, pir_server *pir.SimplePIR, p pir.Params) {
	defer conn.Close()

	/*-------------offline phase-------------*/
	fmt.Println("-------------offline phase start-------------")

	fmt.Println("assuming data has trasfered.")

	fmt.Println("-------------offline phase end-------------")
	/*-------------offline phase-------------*/

	/*--------------online phase-------------*/
	fmt.Println("--------------online phase start-------------")

	// 1. receive query
	var query pir.MsgSlice
	decoder := json.NewDecoder(conn)
	err := decoder.Decode(&query)
	if err != nil {
		fmt.Println("Error decoding JSON:", err.Error())
		return
	}
	fmt.Println("1. Receive query.")

	// pack DB
	pir_server.PackDB(DB, p)

	// 2. answer query
	answer := pir_server.Answer(DB, query, server_state, shared_state, p) // ans = DB * qu
	pir_server.Reset(DB, p)                                               // unpack DB
	encoder := json.NewEncoder(conn)
	err = encoder.Encode(answer)
	if err != nil {
		fmt.Println("Error encoding answer:", err.Error())
		return
	}
	fmt.Println("2. Send answer.")

	fmt.Println("--------------online phase end-------------")
	/*--------------online phase-------------*/

}

func main() {
	// load pre computed data
	fmt.Println("--------------pre loading start-------------")
	pc_file_path, err := os.Open("./preprocess/data/pre_computed_data")
	if err != nil {
		fmt.Println("Error opening pre_computed_file:", err)
		return
	}
	decoder := json.NewDecoder(pc_file_path)
	var pre_computed_data config.Pre_computed_data
	err = decoder.Decode(&pre_computed_data)
	if err != nil {
		fmt.Println("Error decoding pre_computed_data:", err)
		return
	}
	pc_file_path.Close()
	fmt.Println("--------------pre loading end-------------")

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

		go handleConnection(conn, pre_computed_data.DB, pre_computed_data.Server_state, pre_computed_data.Shared_state, &pre_computed_data.Pir_server, pre_computed_data.P)
	}
}
