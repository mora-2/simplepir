package pir

import (
	"fmt"
	"math"
)

type State struct {
	Data []*Matrix
}

type CompressedState struct {
	Seed *PRGKey
}

type Msg struct {
	Data []*Matrix
}

func (m *Msg) Size() uint64 {
	sz := uint64(0)
	for _, d := range m.Data {
		sz += d.Size()
	}
	return sz
}

type MsgSlice struct {
	Data []Msg
}

func (m *MsgSlice) Size() uint64 {
	sz := uint64(0)
	for _, d := range m.Data {
		sz += d.Size()
	}
	return sz
}

func MakeState(elems ...*Matrix) State {
	st := State{}
	for _, elem := range elems {
		st.Data = append(st.Data, elem)
	}
	return st
}

func MakeCompressedState(elem *PRGKey) CompressedState {
	st := CompressedState{}
	st.Seed = elem
	return st
}

func MakeMsg(elems ...*Matrix) Msg {
	msg := Msg{}
	for _, elem := range elems {
		msg.Data = append(msg.Data, elem)
	}
	return msg
}

func MakeMsgSlice(elems ...Msg) MsgSlice {
	slice := MsgSlice{}
	for _, elem := range elems {
		slice.Data = append(slice.Data, elem)
	}
	return slice
}

// Returns the i-th elem in the representation of m in base p.
func Base_p(p, m, i uint64) uint64 {
	for j := uint64(0); j < i; j++ {
		m = m / p
	}
	return (m % p)
}

// Returns the element whose base-p decomposition is given by the values in vals
func Reconstruct_from_base_p(p uint64, vals []uint64) uint64 {
	res := uint64(0)
	coeff := uint64(1)
	for _, v := range vals {
		res += coeff * v
		coeff *= p
	}
	return res
}

// Returns how many entries in Z_p are needed to represent an element in Z_q
func Compute_num_entries_base_p(p, log_q uint64) uint64 {
	log_p := math.Log2(float64(p))
	return uint64(math.Ceil(float64(log_q) / log_p))
}

// Returns how many Z_p elements are needed to represent a database of N entries,
// each consisting of row_length bits.
func Num_DB_entries(N, row_length, p uint64) (uint64, uint64, uint64) {
	if float64(row_length) <= math.Log2(float64(p)) {
		// pack multiple DB entries into a single Z_p elem
		logp := uint64(math.Log2(float64(p)))
		entries_per_elem := logp / row_length
		db_entries := uint64(math.Ceil(float64(N) / float64(entries_per_elem)))
		if db_entries == 0 || db_entries > N {
			fmt.Printf("Num entries is %d; N is %d\n", db_entries, N)
			panic("Should not happen")
		}
		return db_entries, 1, entries_per_elem
	}

	// use multiple Z_p elems to represent a single DB entry
	ne := Compute_num_entries_base_p(p, row_length)
	return N * ne, ne, 0
}

func Num_StrDB_entries(N, row_length, p uint64, Item_bits uint64) (uint64, uint64, uint64) {
	if float64(row_length) <= math.Log2(float64(p)) {
		// pack multiple DB entries into a single Z_p elem
		logp := uint64(math.Log2(float64(p)))
		entries_per_elem := logp / row_length
		db_entries := uint64(math.Ceil(float64(N) / float64(entries_per_elem)))
		if db_entries == 0 || db_entries > N {
			fmt.Printf("Num entries is %d; N is %d\n", db_entries, N)
			panic("Should not happen")
		}
		return db_entries, 1, entries_per_elem
	}

	// use multiple Z_p elems to represent a single DB entry
	ne := Compute_num_str_entries_base_p(p, row_length, Item_bits)
	return N * ne, ne, 0
}

// Returns how many entries in Z_p are needed to represent an element in Z_q
func Compute_num_str_entries_base_p(p, log_q uint64, Item_bits uint64) uint64 {
	log_p := math.Log2(float64(p))
	if log_q > 64 {
		return uint64(math.Ceil(float64(64)/log_p) * math.Ceil(float64(log_q)/64))
	}
	return uint64(math.Ceil(float64(log_q) / log_p))
}

func avg(data []float64) float64 {
	sum := 0.0
	num := 0.0
	for _, elem := range data {
		sum += elem
		num += 1.0
	}
	return sum / num
}

func stddev(data []float64) float64 {
	avg := avg(data)
	sum := 0.0
	num := 0.0
	for _, elem := range data {
		sum += math.Pow(elem-avg, 2)
		num += 1.0
	}
	variance := sum / num // not -1!
	return math.Sqrt(variance)
}

func stringToASCIIArray(s string) []uint64 {
	var results []uint64
	var result uint64
	flag := 0
	for _, char := range s {
		if flag == 8 {
			results = append(results, result)
			result = 0
			flag = 0
		}
		flag += 1
		result = result*256 + uint64(char)
	}
	if result != 0 {
		results = append(results, result)
	}
	return results
}

func intTostring(us []uint64) string {
	var result string
	for i := len(us) - 1; i >= 0; i-- {
		for us[i]/256 != 0 {
			result = string(byte(us[i]%256)) + result
			us[i] = us[i] / 256
		}
		if us[i]%256 != 0 {
			result = string(byte(us[i]%256)) + result
		}
	}
	return result
}

func FindLongestElement(slice []string) string {
	if len(slice) == 0 {
		return "" // 切片为空时，返回空字符串或者你认为合适的默认值
	}

	longest := slice[0]
	for _, str := range slice {
		if len(str) > len(longest) {
			longest = str
		}
	}
	return longest
}
