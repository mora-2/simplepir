// server.go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/mora-2/simplepir/http/server/config"
	"github.com/mora-2/simplepir/pir"
)

var program = "server.go"
var ip_config_file_path string = "./config/ip_config.json"
var pre_comp_file_path string = "./data/pre_computed_data"
var log_file_path string = "log.txt"

func handleConnection(conn net.Conn, DB *pir.Database, server_state pir.State, shared_state pir.State, pir_server *pir.SimplePIR, p pir.Params) {
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

	/*--------------online phase-------------*/
	start := time.Now()
	// 1. receive query
	var query pir.MsgSlice
	decoder := json.NewDecoder(conn)
	err := decoder.Decode(&query)
	if err != nil {
		fmt.Println("Error decoding JSON:", err.Error())
		return
	}
	// fmt.Println("1. Receive query.")
	log.Printf("[%v][%v][1. Receive query]\t Elapsed:%v\t Size:%vKB", program, tcpAddr.String(),
		time.Since(start), query.Size()*uint64(p.Logq)/(8.0*1024.0))

	// 2. answer query
	start = time.Now()
	answer := pir_server.Answer(DB, query, server_state, shared_state, p) // ans = DB * qu
	encoder := json.NewEncoder(conn)
	err = encoder.Encode(answer)
	if err != nil {
		fmt.Println("Error encoding answer:", err.Error())
		return
	}
	// fmt.Println("2. Send answer.")
	log.Printf("[%v][%v][2. Send answer]\t Elapsed:%v\t Size:%vKB", program, tcpAddr.String(),
		time.Since(start), answer.Size()*uint64(p.Logq)/(8.0*1024.0))

	/*--------------online phase-------------*/
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

	// load pre computed data
	/*--------------pre loading start-------------*/
	fmt.Printf("\rLoading...")
	start := time.Now()
	pc_file, err := os.Open(pre_comp_file_path)
	if err != nil {
		fmt.Println("Error opening pre_computed_file:", err.Error())
		return
	}
	decoder = json.NewDecoder(pc_file)
	var pre_computed_data config.Pre_computed_data
	err = decoder.Decode(&pre_computed_data)
	if err != nil {
		fmt.Println("Error decoding pre_computed_data:", err.Error())
		return
	}
	pc_file.Close()

	// pack DB
	pre_computed_data.Pir_server.PackDB(pre_computed_data.DB, pre_computed_data.P)

	log.Printf("[%v][Pre loading]\t Elapsed:%v\t", program, time.Since(start))
	fmt.Printf("\rData loaded.\n")
	/*--------------pre loading end-------------*/

	// start listening
	listener, err := net.Listen("tcp", ":"+fmt.Sprint(ip_cfg.OnlinePort))
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		return
	}
	defer listener.Close()
	fmt.Println("Server listening on :" + fmt.Sprint(ip_cfg.OnlinePort))

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err.Error())
			break
		}

		go handleConnection(conn, pre_computed_data.DB, pre_computed_data.Server_state, pre_computed_data.Shared_state, &pre_computed_data.Pir_server, pre_computed_data.P)
	}
	pre_computed_data.Pir_server.Reset(pre_computed_data.DB, pre_computed_data.P) // unpack DB
}
