// server_offline.go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/mora-2/simplepir/http/server/config"
)

var program = "server_offline.go"
var ip_config_file_path string = "../config/ip_config.json"
var shared_data_file_path string = "../data/shared_data"
var log_file_path string = "log.txt"

func handleConnection(conn net.Conn, shared_data config.Shared_data) {
	defer conn.Close()
	// get connect ip_addr & port
	remoteAddr := conn.RemoteAddr()
	tcpAddr, ok := remoteAddr.(*net.TCPAddr)
	if !ok {
		fmt.Println("Error getting TCPAddr from RemoteAddr")
		conn.Close()
		return
	}
	fmt.Printf("Connection from %s start.\n", tcpAddr.String())

	// time
	start := time.Now()

	encoder := json.NewEncoder(conn)
	err := encoder.Encode(shared_data)
	if err != nil {
		fmt.Println("Error encoding shared_data:", err.Error())
	}
	// fmt.Println("Offline shared_data sent.")
	log.Printf("[%v][%v][Sent answer]\t Elapsed:%v", program, tcpAddr.String(), time.Since(start))

	fmt.Printf("Connection from %s closed.\n", tcpAddr.String())
}

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

	//create log file
	logFile, err := os.Create(log_file_path)
	if err != nil {
		log.Fatal("Cannot create log file: ", err.Error())
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	// load file
	fmt.Printf("\rLoading...")
	shared_file, err := os.Open(shared_data_file_path)
	if err != nil {
		fmt.Println("Error opening shared_data_file:", err.Error())
	}
	defer shared_file.Close()

	decoder = json.NewDecoder(shared_file)
	var shared_data config.Shared_data
	err = decoder.Decode(&shared_data)
	if err != nil {
		fmt.Println("Error loading shared_data:", err.Error())
	}
	fmt.Printf("\rData loaded.\n")

	// start listening
	listener, err := net.Listen("tcp", "219.245.186.51"+":"+fmt.Sprint(ip_cfg.OfflinePort))
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		return
	}
	defer listener.Close()
	fmt.Println("Server listening on :" + fmt.Sprint(ip_cfg.OfflinePort))

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err.Error())
			break
		}

		go handleConnection(conn, shared_data)
	}
}
