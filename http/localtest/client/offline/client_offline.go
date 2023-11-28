// client_offline.go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/mora-2/simplepir/http/localtest/client/config"
	"github.com/mora-2/simplepir/pir"
)

var ip_config_file_path string = "../config/ip_config.json"
var offline_file_path string = "../data/offline_data"

func main() {
	ip_file, err := os.Open(ip_config_file_path)
	if err != nil {
		fmt.Println("Error loading ip_config.json:", err.Error())
	}
	defer ip_file.Close()

	var ip_cfg config.IP_Conn
	decoder := json.NewDecoder(ip_file)
	err = decoder.Decode(&ip_cfg)
	if err != nil {
		fmt.Println("Error decoding ip_config:", err.Error())
	}

	ip_addr := ip_cfg.Ip + ":" + fmt.Sprint(ip_cfg.Port)
	conn, err := net.Dial("tcp", ip_addr)
	if err != nil {
		fmt.Println("Error connecting:", err.Error())
	}
	defer conn.Close()

	/*-------------offline phase-------------*/
	fmt.Println("-------------offline phase start-------------")

	var shared_data config.Shared_data
	decoder = json.NewDecoder(conn)
	err = decoder.Decode(&shared_data)
	if err != nil {
		fmt.Println("Error receiving shared_data:", err.Error())
	}
	fmt.Println("1. Receive shared_data.")

	// start time
	start := time.Now()

	// create client_pir
	client_pir := pir.SimplePIR{}
	// decompress shared_data
	shared_state := client_pir.DecompressState(shared_data.Info, shared_data.P, shared_data.Comp)

	// log out
	log.Printf("Offline phase        \t\tElapsed:%v \tOffline download(hint):%vKB", pir.PrintTime(start),
		float64(shared_data.Offline_download.Size()*uint64(shared_data.P.Logq)/(8.0*1024.0)))
	fmt.Println("-------------offline phase end-------------")
	/*-------------offline phase-------------*/

	offline_data := config.Offline_data{
		Info:             shared_data.Info,
		P:                shared_data.P,
		Shared_state:     shared_state,
		Offline_download: shared_data.Offline_download,
	}

	offline_data_file, err := os.Create(offline_file_path)
	if err != nil {
		fmt.Println("Error creating offline_data_file:", err.Error())
	}
	defer func() {
		if cerr := offline_data_file.Close(); cerr != nil {
			fmt.Println("Error closing offline_data_file:", cerr.Error())

		}
	}()
	encoder := json.NewEncoder(offline_data_file)
	err = encoder.Encode(offline_data)
	if err != nil {
		fmt.Println("Error encoding offline_data_file:", err.Error())
	}

}
