// client.go
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/mora-2/simplepir/http/client/config"
	"github.com/mora-2/simplepir/pir"
)

var program string = "client.go"
var ip_config_file_path string = "./config/ip_config.json"
var offline_file_path string = "./data/offline_data"
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

	ip_addr := ip_cfg.IpAddr + ":" + fmt.Sprint(ip_cfg.OnlinePort)
	conn, err := net.Dial("tcp", ip_addr)
	if err != nil {
		fmt.Println("Error connecting:", err.Error())
		return
	}
	defer conn.Close()

	//create log file
	logFile, err := os.OpenFile(log_file_path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal("Cannot create log file: ", err.Error())
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	/*--------------pre loading start-------------*/
	fmt.Printf("\rLoading...")
	offline_file, err := os.Open(offline_file_path)
	if err != nil {
		fmt.Println("Error opening offline_data file:", err.Error())
	}
	defer offline_file.Close()

	var offline_data config.Offline_data
	decoder = json.NewDecoder(offline_file)
	err = decoder.Decode(&offline_data)
	if err != nil {
		fmt.Println("Error loading offline_data:", err.Error())
	}

	// create client_pir
	client_pir := pir.SimplePIR{}
	fmt.Printf("\rData loaded.\n")
	/*--------------pre loading end-------------*/

	/*--------------online phase-------------*/
	fmt.Println("--------------online phase start-------------")

	// Scan query_index
	query_index := []uint64{}

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Printf("Input the indexes you want to query([0-%v]):", offline_data.Info.Num-1)
	scanner.Scan()
	input_str := scanner.Text()
	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading Input the indexes:", err.Error())
		return
	}

	input_slice := strings.Fields(input_str)
	for _, numStr := range input_slice { // convert str into uint64
		num, err := strconv.ParseUint(numStr, 10, 64)
		if err != nil {
			fmt.Printf("Error converting %s to uint64: %v\n", numStr, err.Error())
			continue
		}
		query_index = append(query_index, num)
	}

	// 1. build query
	start := time.Now()
	var client_state []pir.State // holding secrets
	var query pir.MsgSlice       // holding queries
	for index, _ := range query_index {
		// index_to_query := i[index] + uint64(index)*batch_sz
		index_to_query := query_index[index]
		cs, q := client_pir.Query(index_to_query, offline_data.Shared_state, offline_data.P, offline_data.Info) // 依次制作query语句，qu = As + e + ΔUi
		client_state = append(client_state, cs)
		query.Data = append(query.Data, q)
	}

	encoder := json.NewEncoder(conn)
	err = encoder.Encode(query)
	if err != nil {
		fmt.Println("Error encoding query:", err.Error())
		return
	}
	fmt.Println("1. Send built query.")
	fmt.Printf("\tquery.Data[0].Data[0].Rows: %v\n", query.Data[0].Data[0].Rows)
	fmt.Printf("\tquery.Data[0].Data[0].Cols: %v\n", query.Data[0].Data[0].Cols)
	fmt.Printf("\tquery.Data[0].Data[0].Data[:5]: %v\n", query.Data[0].Data[0].Data[:5])

	// log out
	log.Printf("[%v][%v][1. Send built query]\t Elapsed:%v \tSize:%vKB", program, conn.LocalAddr(),
		pir.PrintTime(start), float64(query.Size()*uint64(offline_data.P.Logq)/(8.0*1024.0)))

	// 2. Receive answer
	start = time.Now()
	var answer pir.Msg
	decoder = json.NewDecoder(conn)
	err = decoder.Decode(&answer)
	if err != nil {
		fmt.Println("Error decoding answer:", err.Error())
		return
	}
	fmt.Println("2. Receive answer.")
	fmt.Printf("\tanswer.Data[0].Rows: %v\n", answer.Data[0].Rows)
	fmt.Printf("\tanswer.Data[0].Cols: %v\n", answer.Data[0].Cols)
	fmt.Printf("\tanswer.Data[0].Data[:5]: %v\n", answer.Data[0].Data[:5])

	// log out
	log.Printf("[%v][%v][2. Receive answer]\t Elapsed:%v \tSize:%vKB", program, conn.LocalAddr(),
		pir.PrintTime(start), float64(answer.Size()*uint64(offline_data.P.Logq)/(8.0*1024.0)))

	// 3. Resconstruction
	start = time.Now()
	var result []string
	for index, _ := range query_index {
		// index_to_query := i[index] + uint64(index)*batch_sz
		index_to_query := query_index[index]
		val := client_pir.StrRecover(index_to_query, uint64(index), offline_data.Offline_download,
			query.Data[index], answer, offline_data.Shared_state,
			client_state[index], offline_data.P, offline_data.Info) // 返回指定下标的元素
		result = append(result, val)
	}
	fmt.Println("3. Resconstruction finished.")
	fmt.Printf("\tSimplePIR result: %v\n", result)

	// log out
	log.Printf("[%v][%v][3. Resconstruction]\t Elapsed:%v", program, conn.LocalAddr(), pir.PrintTime(start))

	fmt.Println("--------------online phase end-------------")
	/*--------------online phase-------------*/
}
