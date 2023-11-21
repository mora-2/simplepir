package pir

import (
	"encoding/binary"
	"fmt"
	"os"
)

func my_load_db(bf_int32bin_path string) []uint64 {
	// 打开二进制文件
	file, err := os.Open(bf_int32bin_path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// 计算文件大小
	fileInfo, err := file.Stat()
	if err != nil {
		panic(err)
	}
	fileSize := fileInfo.Size()

	// 创建切片来存储读取的数据
	data := make([]uint64, fileSize/4)

	// 创建缓冲区
	buf := make([]byte, 4)

	// 循环读取二进制文件中的数据
	for i := 0; i < len(data); i++ {
		_, err := file.Read(buf)
		if err != nil {
			panic(err)
		}

		// 将读取的数据转换成int32类型的数字
		num := int32(binary.LittleEndian.Uint32(buf))

		// 将int32类型的数字转换成uint64类型并存储到切片中
		data[i] = uint64(num)

	}

	// 输出读取的数据
	// fmt.Println(data)
	// fmt.Println("idx:", 1000, "\tdata", data[1000])
	// fmt.Println("idx:", 860819, "\tdata", data[860819])
	fmt.Printf("DB size:%v\n", len(data))
	return data
}

func MakeMyDB(Num, row_length uint64, p *Params, valslice [][]uint64) *Database {
	D := SetupDB(Num, row_length, p)
	D.Data = MatrixZeros(p.L, p.M)

	if uint64(len(valslice)*len(valslice[0])) != Num {
		panic("Bad input DB")
	}

	if D.Info.Packing > 0 {
		for slice_idx, vals := range valslice {
			// Pack multiple DB elems into each Z_p elem
			at := uint64(0)
			cur := uint64(0)
			coeff := uint64(1)
			for i, elem := range vals {
				cur += (elem * coeff)
				coeff *= (1 << row_length)
				if ((i+1)%int(D.Info.Packing) == 0) || (i == len(vals)-1) {
					D.Data.Set(cur, uint64(slice_idx), at%p.M)
					// fmt.Println("Row:", uint64(slice_idx), "\tCols:", at%p.M, "\tData:", cur)
					at += 1
					cur = 0
					coeff = 1
				}
			}
		}
	} else {
		// // Use multiple Z_p elems to represent each DB elem
		// for i, elem := range vals {
		// 	for j := uint64(0); j < D.Info.Ne; j++ {
		// 		D.Data.Set(Base_p(D.Info.P, elem, j), (uint64(i)/p.M)*D.Info.Ne+j, uint64(i)%p.M)
		// 	}
		// }
		panic("Use multiple Z_p elems to represent each DB elem. Not implemented yet!")
	}

	// // Map DB elems to [-p/2; p/2]
	// D.Data.Sub(p.P / 2)

	return D
}
