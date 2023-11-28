// client_offline.go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/mora-2/simplepir/http/client/config"
	"github.com/mora-2/simplepir/pir"
)

var program string = "client_offline.go"
var ip_config_file_path string = "../config/ip_config.json"
var offline_file_path string = "../data/offline_data"
var log_file_path string = "log.txt"

func main() {
	// ip config
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

	ip_addr := ip_cfg.IpAddr + ":" + fmt.Sprint(ip_cfg.OfflinePort)
	conn, err := net.Dial("tcp", ip_addr)
	if err != nil {
		fmt.Println("Error connecting:", err.Error())
	}
	defer conn.Close()

	//create log file
	logFile, err := os.OpenFile(log_file_path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal("Cannot create log file: ", err.Error())
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	/*-------------offline phase-------------*/
	// start time
	start := time.Now()

	fmt.Printf("\rReceiving...")
	var shared_data config.Shared_data
	decoder = json.NewDecoder(conn)
	err = decoder.Decode(&shared_data)
	if err != nil {
		fmt.Println("Error receiving shared_data:", err.Error())
	}
	// fmt.Println("1.1 Receive shared_data.")
	// log out
	log.Printf("[%v][%v][1.1 Receive shared_data]\t Elapsed:%v", program, conn.LocalAddr(), time.Since(start))
	fmt.Printf("\rData received.\n")

	start = time.Now()
	fmt.Printf("\rDecompressing...")
	// create client_pir
	client_pir := pir.SimplePIR{}
	// decompress shared_data(matrix A)
	shared_state := client_pir.DecompressState(shared_data.Info, shared_data.P, shared_data.Comp)

	// log out
	log.Printf("[%v][%v][1.2 Decompress]\t Elapsed:%v\n", program, conn.LocalAddr(), time.Since(start))
	fmt.Printf("\rData decompressed.\n")

	offline_data := config.Offline_data{
		Info:             shared_data.Info,
		P:                shared_data.P,
		Shared_state:     shared_state,
		Offline_download: shared_data.Offline_download,
	}

	start = time.Now()
	fmt.Printf("\rWriting...")
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
	// log out
	log.Printf("[%v][%v][2 Write]\t Elapsed:%v\n", program, conn.LocalAddr(), time.Since(start))
	fmt.Printf("\rData wrote.\n")
	/*-------------offline phase-------------*/
}
