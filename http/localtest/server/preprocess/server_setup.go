// server_setup.go
package main

import (
	"encoding/json"
	"fmt"
	"os"

	// "runtime"
	// "time"

	"github.com/mora-2/simplepir/http/localtest/server/config"
	"github.com/mora-2/simplepir/pir"
)

func main() {
	// server params config
	const LOGQ = uint64(32)           // ciphertext mod
	const SEC_PARAM = uint64(1 << 10) // secret demension

	// db_vals := []uint64{141, 13, 52, 43, 44}
	//db_vals := []string{"apple", "banana", "cat", "dog"}
	data_file_path := "../../../../data/data.csv"
	db_vals := pir.LoadFile(data_file_path, "Child's First Name")
	N := uint64(len(db_vals))
	d := uint64(len(pir.FindLongestElement(db_vals)) * 8) // sizeof(byte): 8
	pir_server := pir.SimplePIR{}
	p := pir_server.PickStrParams(N, d, SEC_PARAM, LOGQ)

	// DB loading
	DB := pir.MakeStrDB(N, d, &p, db_vals)

	// create A, H
	shared_state, shared_comp := pir_server.InitCompressed(DB.Info, p)
	server_state, offline_download := pir_server.SetupUnpackedDB(DB, shared_state, p) // 计算H矩阵，并将DB元素映射到[0，p]

	shared_data := config.Shared_data{
		Info:             DB.Info,
		P:                p,
		Comp:             shared_comp,
		Offline_download: offline_download,
	}
	pre_computed_data := config.Pre_computed_data{
		DB:           DB,
		Server_state: server_state,
		Shared_state: shared_state,
		Pir_server:   pir_server,
		P:            p,
	}

	// offline: send shared_data
	shared_data_file, err := os.Create("../data/shared_data")
	if err != nil {
		fmt.Println("Error creating shared_data_file:", err.Error())
	}
	defer func() {
		if cerr := shared_data_file.Close(); cerr != nil {
			fmt.Println("Error closing shared_data_file:", cerr.Error())

		}
	}()
	encoder := json.NewEncoder(shared_data_file)
	err = encoder.Encode(shared_data)
	if err != nil {
		fmt.Println("Error encoding shared_data_file:", err.Error())
	}

	// save pre_computed_data
	pre_computed_data_file, err := os.Create("../data/pre_computed_data")
	if err != nil {
		fmt.Println("Error creating pre_computed_data_file:", err.Error())
	}
	defer func() {
		if cerr := pre_computed_data_file.Close(); cerr != nil {
			fmt.Println("Error closing pre_computed_data:", cerr.Error())

		}
	}()
	encoder = json.NewEncoder(pre_computed_data_file)
	err = encoder.Encode(pre_computed_data)
	if err != nil {
		fmt.Println("Error encoding pre_computed_data_file:", err.Error())
	}

}
