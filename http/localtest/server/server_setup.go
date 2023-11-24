package main

import (
	"encoding/json"
	"fmt"
	//"net"
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

func main() {
	fmt.Println("--------------pre loading-------------")
	// server params config
	const LOGQ = uint64(32)           // ciphertext mod
	const SEC_PARAM = uint64(1 << 10) // secret demension

	// db_vals := []uint64{141, 13, 52, 43, 44}
	//db_vals := []string{"apple", "banana", "cat", "dog"}
	filepath := "../../../data/data.csv"
	db_vals := pir.LoadFile(filepath, "Child's First Name")
	N := uint64(len(db_vals))
	d := uint64(len(pir.FindLongestElement(db_vals)) * 8) // sizeof(byte): 8
	pir_server := pir.SimplePIR{}
	p := pir_server.PickStrParams(N, d, SEC_PARAM, LOGQ)

	// DB loading
	DB := pir.MakeStrDB(N, d, &p, db_vals)

	// start listening
	shared_state, shared_comp := pir_server.InitCompressed(DB.Info, p)
	server_state, offline_download := pir_server.SetupUnpackedDB(DB, shared_state, p) // 计算H矩阵，并将DB元素映射到[0，p]
	shared_data := shared_data{
		Info:             DB.Info,
		P:                p,
		Comp:             shared_comp,
		Offline_download: offline_download,
	}
	func_data := func_data{
		DB: 			DB,
		Server_state:	server_state,
		Shared_state:	shared_state,
		Pir_server:		pir_server,
		P:				p,
	}
	file0, err0 := os.Create("shared_data")
	defer file0.Close()
	if err0 != nil {
		panic(err0)
	}
	encoder0 := json.NewEncoder(file0)
	err0 = encoder0.Encode(shared_data)
	if err0 != nil {
		panic(err0)
	}
	file, err := os.Create("func_data")
	defer file.Close()
	if err != nil {
		panic(err)
	}
	encoder := json.NewEncoder(file)
	err = encoder.Encode(func_data)
	if err != nil {
		panic(err)
	}
	
}
