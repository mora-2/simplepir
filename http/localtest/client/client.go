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

	"github.com/mora-2/simplepir/http/localtest/server/config"
	"github.com/mora-2/simplepir/pir"
)

var shared_file_path string = "./data/shared_data"

func main() {
	conn, err := net.Dial("tcp", "localhost:8080")
	// conn, err := net.Dial("tcp", "219.245.186.116:8080")
	if err != nil {
		fmt.Println("Error connecting:", err.Error())
		return
	}
	defer conn.Close()

	//create log file
	logFile, err := os.Create("log.txt")
	if err != nil {
		log.Fatal("Cannot create log file: ", err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	// create client_pir
	client_pir := pir.SimplePIR{}

	/*-------------offline phase-------------*/
	fmt.Println("-------------offline phase start-------------")

	// start time
	start := time.Now()
	// 1. get shared_data from server response
	var shared_data config.Shared_data
	shared_file, err := os.Open(shared_file_path) // assume the file has been downloaded
	if err != nil {
		fmt.Println("Error opening shared_file:", err)
		return
	}
	decoder := json.NewDecoder(shared_file)
	err = decoder.Decode(&shared_data)
	if err != nil {
		fmt.Println("Error decoding shared_data:", err.Error())
		return
	}
	fmt.Println("1. Receive shared_data.")

	// decompress shared_data
	shared_state := client_pir.DecompressState(shared_data.Info, shared_data.P, shared_data.Comp)

	// log out
	log.Printf("Offline phase        \t\tElapsed:%v \tOffline download(hint):%vKB", pir.PrintTime(start),
		float64(shared_data.Offline_download.Size()*uint64(shared_data.P.Logq)/(8.0*1024.0)))
	fmt.Println("-------------offline phase end-------------")
	/*-------------offline phase-------------*/

	/*--------------online phase-------------*/
	fmt.Println("--------------online phase start-------------")

	// Scan query_index
	query_index := []uint64{}

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Printf("Input the indexes you want to query([0-%v]):", shared_data.Info.Num-1)
	scanner.Scan()
	input_str := scanner.Text()
	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading Input the indexes:", err)
		return
	}

	input_slice := strings.Fields(input_str)
	for _, numStr := range input_slice { // convert str into uint64
		num, err := strconv.ParseUint(numStr, 10, 64)
		if err != nil {
			fmt.Printf("Error converting %s to uint64: %v\n", numStr, err)
			continue
		}
		query_index = append(query_index, num)
	}

	// 1. build query
	start = time.Now()
	var client_state []pir.State // holding secrets
	var query pir.MsgSlice       // holding queries
	for index, _ := range query_index {
		// index_to_query := i[index] + uint64(index)*batch_sz
		index_to_query := query_index[index]
		cs, q := client_pir.Query(index_to_query, shared_state, shared_data.P, shared_data.Info) // 依次制作query语句，qu = As + e + ΔUi
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
	log.Printf("Online phase(1. Build query)    \tElapsed:%v   \tupload:%vKB", pir.PrintTime(start),
		float64(query.Size()*uint64(shared_data.P.Logq)/(8.0*1024.0)))

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
	log.Printf("Online phase(2. Receive answer) \tElapsed:%v \tdownloadload:%vKB", pir.PrintTime(start),
		float64(answer.Size()*uint64(shared_data.P.Logq)/(8.0*1024.0)))

	// 3. Resconstruction
	start = time.Now()
	var result []string
	for index, _ := range query_index {
		// index_to_query := i[index] + uint64(index)*batch_sz
		index_to_query := query_index[index]
		val := client_pir.StrRecover(index_to_query, uint64(index), shared_data.Offline_download,
			query.Data[index], answer, shared_state,
			client_state[index], shared_data.P, shared_data.Info) // 返回指定下标的元素
		result = append(result, val)
	}
	fmt.Println("3. Resconstruction finished.")
	fmt.Printf("\tSimplePIR result: %v\n", result)

	// log out
	log.Printf("Online phase(3. Resconstruction) \tElapsed:%v", pir.PrintTime(start))

	fmt.Println("--------------online phase end-------------")
	/*--------------online phase-------------*/
}
