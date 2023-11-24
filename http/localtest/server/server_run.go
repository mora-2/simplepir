// server.go
package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
//	"encoding/gob"
	// "runtime"
	// "time"

	"github.com/mora-2/simplepir/pir"
)

type shared_data struct {
	Info             pir.DBinfo
	P                pir.Params
	Comp             pir.CompressedState
	Offline_download pir.Msg
}
type func_data struct{
	DB         		*pir.Database
	Server_state 	pir.State
	Shared_state	pir.State
	Pir_server		pir.SimplePIR
	P				pir.Params
}

func handleConnection(conn net.Conn, DB *pir.Database,server_state pir.State,shared_state pir.State, pir_server *pir.SimplePIR, p pir.Params) {
	defer conn.Close()

	{
		_ = server_state
	}
	fmt.Println("-------------offline pahse-------------")
	/*-------------offline pahse-------------*/

	/*--------------online pahse-------------*/
	fmt.Println("--------------online pahse-------------")

	// 1. receive query
	var query pir.MsgSlice
	decoder := json.NewDecoder(conn)
	err := decoder.Decode(&query)
	if err != nil {
		fmt.Println("Error decoding JSON:", err.Error())
		return
	}
	fmt.Println("1. Receice query.")

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
	// fmt.Printf("answer.Data[0]: %v\n", answer.Data[0])

	fmt.Println("--------------online pahse-------------")
	/*--------------online pahse-------------*/

	// // 3. Answer check
	// query_index := []uint64{}
	// decoder = json.NewDecoder(conn)
	// err = decoder.Decode(&query_index)
	// if err != nil {
	// 	fmt.Println("Error decoding query_index:", err.Error())
	// 	return
	// }

	// result := []uint64{}
	// for _, value := range query_index { // get elem in DB
	// 	result = append(result, DB.GetElem(value))
	// }
	// fmt.Printf("SimplePIR result in DB:%+v\n", result)
}

func main() {
	file, err0 := os.Open("func_data")
    if err0 != nil {
        fmt.Println(err0)
        return
    }
    decoder := json.NewDecoder(file)
    var func_data func_data
    err0 = decoder.Decode(&func_data)
    if err0 != nil {
        fmt.Println(err0)
        return
    }
    file.Close()
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

		go handleConnection(conn, func_data.DB,func_data.Server_state,func_data.Shared_state, &func_data.Pir_server, func_data.P)
	}
}
