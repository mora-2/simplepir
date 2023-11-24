// server.go
package main

import (
	"encoding/json"
	"fmt"
	"net"

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

func handleConnection(conn net.Conn, DB *pir.Database, pir_server *pir.SimplePIR, p pir.Params) {
	defer conn.Close()
	/*-------------offline pahse-------------*/
	fmt.Println("-------------offline pahse-------------")

	// 1. send shared_data help client construct matrix A
	shared_state, shared_comp := pir_server.InitCompressed(DB.Info, p)
	server_state, offline_download := pir_server.SetupUnpackedDB(DB, shared_state, p) // 计算H矩阵，并将DB元素映射到[0，p]
	shared_data := shared_data{
		Info:             DB.Info,
		P:                p,
		Comp:             shared_comp,
		Offline_download: offline_download,
	}

	encoder := json.NewEncoder(conn)
	err := encoder.Encode(shared_data)
	if err != nil {
		fmt.Println("Error encoding shared_data:", err.Error())
		return
	}
	fmt.Println("1. Send shared_data.")
	// fmt.Printf("shared_state.Data[0]: %v\n", shared_state.Data[0])

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
	err = decoder.Decode(&query)
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
	encoder = json.NewEncoder(conn)
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
	fmt.Println("--------------pre loading-------------")
	// server params config
	const LOGQ = uint64(32)           // ciphertext mod
	const SEC_PARAM = uint64(1 << 10) // secret demension

	// db_vals := []uint64{141, 13, 52, 43, 44}
	db_vals := []string{"apple", "banana", "cat", "dog"}
	N := uint64(len(db_vals))
	d := uint64(len(pir.FindLongestElement(db_vals)) * 8) // sizeof(byte): 8
	pir_server := pir.SimplePIR{}
	p := pir_server.PickParams(N, d, SEC_PARAM, LOGQ)

	// DB loading
	DB := pir.MakeStrDB(N, d, &p, db_vals)

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

		go handleConnection(conn, DB, &pir_server, p)
	}
}
