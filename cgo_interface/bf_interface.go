package main

// #cgo CXXFLAGS: -std=c++11
// #cgo CFLAGS: -I/home/yuance/Work/Encryption/SBT/interface
// #cgo LDFLAGS: -L. -lloadDB
// #include "Interface.h"
import "C"
import "fmt"

// type cgomyLoadDB_T C.myLoadDB_T

func get_bf_data() []uint8 {
	infile_list_path := "/home/yuance/Work/Encryption/PIR/code/PIR/simplepir/cgo_interface/file_path_list.txt"
	hashes_file := "/home/yuance/Work/Encryption/PIR/code/PIR/simplepir/cgo_interface/data/hashfile_k20"
	out_dir := "/home/yuance/Work/Encryption/PIR/code/PIR/simplepir/cgo_interface/out"

	p := C.NewMyLoadDB(infile_list_path, hashes_file, out_dir)
	defer C.DeleteMyLoadDB(p)

	return []uint8([]byte(C.GoStringN(C.Get_Datavector(p, 0), C.Num_BFx_BV_Size(p, 0))))
	// return []uint8{1}
}

func main() {
	v := get_bf_data()
	fmt.Println(v)
}
