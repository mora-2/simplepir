// client.go
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/mora-2/simplepir/http/localtest/server/config"
	"github.com/mora-2/simplepir/pir"
)

var shared_file_path string = "./data/shared_data"

func main() {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Println("Error connecting:", err.Error())
		return
	}
	defer conn.Close()

	// create client_pir
	client_pir := pir.SimplePIR{}

	/*-------------offline pahse-------------*/
	fmt.Println("-------------offline pahse start-------------")

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
	// fmt.Printf("shared_state.Data[0]: %v\n", shared_state.Data[0])

	fmt.Println("-------------offline pahse end-------------")
	/*-------------offline pahse-------------*/

	/*--------------online pahse-------------*/
	fmt.Println("--------------online pahse start-------------")

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
	// fmt.Printf("query_index: %v\n", query_index)

	// 1. build query
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

	// 2. Receive answer
	var answer pir.Msg
	decoder = json.NewDecoder(conn)
	err = decoder.Decode(&answer)
	if err != nil {
		fmt.Println("Error decoding answer:", err.Error())
		return
	}
	fmt.Println("2. Receive answer.")
	// fmt.Printf("answer.Data[0]: %v\n", answer.Data[0])

	// 3. Resconstruction
	var result []string
	for index, _ := range query_index {
		// index_to_query := i[index] + uint64(index)*batch_sz
		index_to_query := query_index[index]
		val := client_pir.StrRecover(index_to_query, uint64(index), shared_data.Offline_download,
			query.Data[index], answer, shared_state,
			client_state[index], shared_data.P, shared_data.Info) // 返回指定下标的元素
		result = append(result, val)
		// if DB.GetElem(index_to_query) != val {
		// 	fmt.Printf("Batch %d (querying index %d -- row should be >= %d): Got %d instead of %d\n",
		// 		index, index_to_query, DB.Data.Rows/4, val, DB.GetElem(index_to_query))
		// 	panic("Reconstruct failed!")
		// }
		// fmt.Println("Simple PIR ---- Query Element Index:", index_to_query, "\tElement in Database:", DB.GetElem(index_to_query), "\tGet Element:", val)
	}
	fmt.Println("3. Resconstruction finished.")
	fmt.Printf("SimplePIR result: %v\n", result)

	fmt.Println("--------------online pahse end-------------")
	/*--------------online pahse-------------*/

	// // 4. send query_index
	// encoder = json.NewEncoder(conn)
	// err = encoder.Encode(query_index)
	// if err != nil {
	// 	fmt.Println("Error encoding query_index:", err.Error())
	// 	return
	// }
	// fmt.Println("4. send query_index.")
}
