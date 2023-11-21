package pir

// #cgo CFLAGS: -O3 -march=native
// #include "pir.h"
import "C"
import (
	"fmt"
)

type SimplePIR struct{}

func (pi *SimplePIR) Name() string {
	return "SimplePIR"
}

func (pi *SimplePIR) PickParams(N, d, n, logq uint64) Params {
	good_p := Params{}
	found := false

	// Iteratively refine p and DB dims, until find tight values
	for mod_p := uint64(2); ; mod_p += 1 {
		l, m := ApproxSquareDatabaseDims(N, d, mod_p) // 自增p，获得对应m

		p := Params{
			N:    n,
			Logq: logq,
			L:    l,
			M:    m,
		}
		// 根据m选择加密参数
		p.PickParams(false, m)

		// found = true, 如果自增的p，超过了选择的P参数，说明最优的LWE加密参数以及对应的数据库l，m找到
		if p.P < mod_p {
			if !found {
				panic("Error; should not happen")
			}
			good_p.PrintParams()
			return good_p
		}

		good_p = p
		found = true
	}

	panic("Cannot be reached")
	return Params{}
}

func (pi *SimplePIR) PickParamsGivenDimensions(l, m, n, logq uint64) Params {
	p := Params{
		N:    n,
		Logq: logq,
		L:    l,
		M:    m,
	}
	
	p.PickParams(false, m)
	return p
}

// Works for SimplePIR because vertical concatenation doesn't increase
// the number of LWE samples (so don't need to change LWE params)
func (pi *SimplePIR) ConcatDBs(DBs []*Database, p *Params) *Database {
	if len(DBs) == 0 {
		panic("Should not happen")
	}

	if DBs[0].Info.Num != p.L*p.M {
		panic("Not yet implemented")
	}

	rows := DBs[0].Data.Rows
	for j := 1; j < len(DBs); j++ {
		if DBs[j].Data.Rows != rows {
			panic("Bad input")
		}
	}

	D := new(Database)
	D.Data = MatrixZeros(0, 0)
	D.Info = DBs[0].Info
	D.Info.Num *= uint64(len(DBs))
	p.L *= uint64(len(DBs))

	for j := 0; j < len(DBs); j++ {
		D.Data.Concat(DBs[j].Data.SelectRows(0, rows))
	}

	return D
}

func (pi *SimplePIR) GetBW(info DBinfo, p Params) {
	offline_download := float64(p.L*p.N*p.Logq) / (8.0 * 1024.0)
	fmt.Printf("\t\tOffline download: %d KB\n", uint64(offline_download))

	online_upload := float64(p.M*p.Logq) / (8.0 * 1024.0)
	fmt.Printf("\t\tOnline upload: %d KB\n", uint64(online_upload))

	online_download := float64(p.L*p.Logq) / (8.0 * 1024.0)
	fmt.Printf("\t\tOnline download: %d KB\n", uint64(online_download))
}

func (pi *SimplePIR) Init(info DBinfo, p Params) State {
	A := MatrixRand(p.M, p.N, p.Logq, 0)
	return MakeState(A)
}

func (pi *SimplePIR) MyInit(info DBinfo, p Params) State {
	// fmt.Println("simplepir.go MyInit 1")
	A := MatrixRand(p.M*info.Packing, p.N, p.Logq, 0)
	// fmt.Println("simplepir.go MyInit 2")
	return MakeState(A)
}

func (pi *SimplePIR) InitCompressed(info DBinfo, p Params) (State, CompressedState) {
	seed := RandomPRGKey()
	bufPrgReader = NewBufPRG(NewPRG(seed))
	return pi.Init(info, p), MakeCompressedState(seed)
}

func (pi *SimplePIR) DecompressState(info DBinfo, p Params, comp CompressedState) State {
	bufPrgReader = NewBufPRG(NewPRG(comp.Seed))
	return pi.Init(info, p)
}

func (pi *SimplePIR) Setup(DB *Database, shared State, p Params) (State, Msg) {
	A := shared.Data[0]
	H := MatrixMul(DB.Data, A)

	// map the database entries to [0, p] (rather than [-p/1, p/2]) and then
	// pack the database more tightly in memory, because the online computation
	// is memory-bandwidth-bound
	DB.Data.Add(p.P / 2)
	DB.Squish()

	return MakeState(), MakeMsg(H)
}

func (pi *SimplePIR) MySetup(DB *Database, shared State, p Params) (State, Msg) {
	A := shared.Data[0]
	H := MyMatrixMul(DB.Data, DB.Info.Packing, A, uint64(1))
	// fmt.Println("hint:", H.Data[0], H.Data[1], "\t row:", H.Rows, "\t col:", H.Cols)
	// fmt.Println("A:", A.Data[0], A.Data[1], "\t row:", A.Rows, "\t col:", A.Cols)
	// // map the database entries to [0, p] (rather than [-p/1, p/2]) and then
	// // pack the database more tightly in memory, because the online computation
	// // is memory-bandwidth-bound
	// DB.Data.Add(p.P / 2)
	// DB.Squish()

	return MakeState(), MakeMsg(H)
}

func (pi *SimplePIR) FakeSetup(DB *Database, p Params) (State, float64) {
	offline_download := float64(p.L*p.N*uint64(p.Logq)) / (8.0 * 1024.0)
	fmt.Printf("\t\tOffline download: %d KB\n", uint64(offline_download))

	// map the database entries to [0, p] (rather than [-p/1, p/2]) and then
	// pack the database more tightly in memory, because the online computation
	// is memory-bandwidth-bound
	DB.Data.Add(p.P / 2)
	DB.Squish()

	return MakeState(), offline_download
}

func (pi *SimplePIR) Query(i uint64, shared State, p Params, info DBinfo) (State, Msg) {
	A := shared.Data[0]

	secret := MatrixRand(p.N, 1, p.Logq, 0)
	err := MatrixGaussian(p.M, 1)
	query := MatrixMul(A, secret)
	query.MatrixAdd(err)
	if info.Packing > 1 {
		query.Data[(i%(p.M*info.Packing))/info.Packing] += C.Elem(p.Delta())
	} else {
		query.Data[i%(p.M)] += C.Elem(p.Delta())
	}

	// Pad the query to match the dimensions of the compressed DB
	if p.M%info.Squishing != 0 {
		query.AppendZeros(info.Squishing - (p.M % info.Squishing))
	}

	return MakeState(secret), MakeMsg(query)
}

func (pi *SimplePIR) MyQuery(i []uint64, shared State, p Params, info DBinfo) (State, Msg) {
	A := shared.Data[0]

	secret := MatrixRand(p.N, 1, p.Logq, 0)
	err := MatrixGaussian(p.M*info.Packing, 1)
	// fmt.Println("secret:", secret.Data[0], secret.Data[1], "\t row:", secret.Rows, "\t col:", secret.Cols)
	// err := MatrixZeros(p.M*info.Packing, 1)
	// err.Add(uint64(0))
	query := MatrixMul(A, secret)
	// fmt.Println("hint*s:", MyMatrixMul(DB.Data, info.Packing, MatrixMul(A, secret), 1))
	// fmt.Println("hint*s:", secret)
	query.MatrixAdd(err)
	for _, idx := range i {
		query.Data[idx] += C.Elem(p.Delta())
	}

	// fmt.Println("query Rows:", query.Rows, "\tquery Cols:", query.Cols)
	// fmt.Println("query data:", query.Data)
	// fmt.Println("err data:", err.Data)

	// query.MatrixSub(MatrixMul(A, secret))
	// query.MatrixSub(err)
	// fmt.Println("ΔUi data:", query.Data)
	// fmt.Println("ΔDBUi data:", MyMatrixMul(DB.Data, info.Packing, query, uint64(1)))
	return MakeState(secret), MakeMsg(query)
}

func (pi *SimplePIR) Answer(DB *Database, query MsgSlice, server State, shared State, p Params) Msg {
	// ans := new(Matrix)
	// num_queries := uint64(len(query.Data)) // number of queries in the batch of queries
	// batch_sz := DB.Data.Rows / num_queries // how many rows of the database each query in the batch maps to
	// last := uint64(0)

	// // Run SimplePIR's answer routine for each query in the batch
	// for batch, q := range query.Data {
	// 	if batch == int(num_queries-1) {
	// 		batch_sz = DB.Data.Rows - last
	// 	}
	// 	a := MatrixMulVecPacked(DB.Data.SelectRows(last, batch_sz),
	// 		q.Data[0],
	// 		DB.Info.Basis,
	// 		DB.Info.Squishing)
	// 	ans.Concat(a)
	// 	last += batch_sz
	// }

	// return MakeMsg(ans)

	ans_msg := Msg{}
	// Run SimplePIR's answer routine for each query
	for _, q := range query.Data {
		a := MatrixMulVecPacked(DB.Data,
			q.Data[0],
			DB.Info.Basis,
			DB.Info.Squishing)

		ans_msg.Data = append(ans_msg.Data, a)
	}

	return ans_msg
}

func (pi *SimplePIR) MyAnswer(DB *Database, query MsgSlice, server State, shared State, p Params) Msg {

	ans_msg := Msg{}
	// Run SimplePIR's answer routine for one integrated query
	a := MyMatrixMul(DB.Data, DB.Info.Packing, query.Data[0].Data[0], 1)
	// fmt.Println("query.Data[0].Data[0]:", query.Data[0].Data[0].Data)
	// fmt.Println("ans Data:", a.Data)

	ans_msg.Data = append(ans_msg.Data, a)

	return ans_msg
}

func (pi *SimplePIR) Recover(i uint64, batch_index uint64, offline Msg, query Msg, answer Msg,
	shared State, client State, p Params, info DBinfo) uint64 {
	// secret := client.Data[0]
	// H := offline.Data[0]
	// ans := answer.Data[0]
	// ratio := p.P / 2
	// offset := uint64(0)
	// for j := uint64(0); j < p.M; j++ {
	// 	offset += ratio * query.Data[0].Get(j, 0)
	// }
	// offset %= (1 << p.Logq)
	// offset = (1 << p.Logq) - offset

	// interm := MatrixMul(H, secret)
	// fmt.Println("ans.Rows:", ans.Rows, "\tans.Cols:", ans.Cols)
	// fmt.Println("interm.Rows:", interm.Rows, "\tinterm.Cols:", interm.Cols)
	// ans.MatrixSub(interm)

	// var vals []uint64
	// // Recover each Z_p element that makes up the desired database entry

	// row := uint64(0)
	// if info.Packing > 0 {
	// 	row = i / (p.M * info.Packing)
	// } else {
	// 	row = i / (p.M)
	// }
	// for j := row * info.Ne; j < (row+1)*info.Ne; j++ {
	// 	noised := uint64(ans.Data[j]) + offset
	// 	denoised := p.Round(noised)
	// 	vals = append(vals, denoised)
	// 	// fmt.Printf("Reconstructing row %d: %d\n", j, denoised)
	// }
	// ans.MatrixAdd(interm)

	// return ReconstructElem(vals, i, info)

	secret := client.Data[0]
	H := offline.Data[0]
	ratio := p.P / 2
	offset := uint64(0)
	for j := uint64(0); j < p.M; j++ {
		offset += ratio * query.Data[0].Get(j, 0)
	}
	offset %= (1 << p.Logq)
	offset = (1 << p.Logq) - offset

	interm := MatrixMul(H, secret)
	ans := answer.Data[batch_index]
	ans.MatrixSub(interm)

	var vals []uint64
	// Recover each Z_p element that makes up the desired database entry

	row := uint64(0)
	if info.Packing > 0 {
		row = i / (p.M * info.Packing)
	} else {
		row = i / (p.M)
	}
	for j := row * info.Ne; j < (row+1)*info.Ne; j++ {
		noised := uint64(ans.Data[j]) + offset
		denoised := p.Round(noised)
		vals = append(vals, denoised)
		// fmt.Printf("Reconstructing row %d: %d\n", j, denoised)
	}
	ans.MatrixAdd(interm)
	return ReconstructElem(vals, i, info)
}

func (pi *SimplePIR) MyRecover(i []uint64, batch_index uint64, offline Msg, query Msg, answer Msg,
	shared State, client State, p Params, info DBinfo) []uint64 {

	// // 创建一个文件，用于存储日志信息
	// file, err1 := os.OpenFile("logfile.txt", os.O_CREATE|os.O_WRONLY, 0644)
	// if err1 != nil {
	// 	log.Fatal(err1)
	// }
	// defer file.Close()

	// // 创建一个新的日志记录器
	// logger := log.New(file, "", log.LstdFlags)

	// fmt.Println("hint*s:", MyMatrixMul(DB.Data, DB.Info.Packing, MatrixMul(shared.Data[0], secret), uint64(1)))
	// fmt.Println("hint*s:", MyMatrixMul(MyMatrixMul(DB.Data, DB.Info.Packing, shared.Data[0], uint64(1)), 1, secret, 1))
	// logger.Println("DB.Data:", DB.Data)
	// logger.Println()
	// logger.Println()
	// logger.Println("A:", shared.Data[0])
	// logger.Println()
	// logger.Println()
	// logger.Println("S:", secret)
	// logger.Println()
	// logger.Println()
	// logger.Println("DB*A:", MyMatrixMul(DB.Data, DB.Info.Packing, shared.Data[0], uint64(1)))
	// logger.Println()
	// logger.Println()
	// mymatrix := &Matrix{18, 2, []C.Elem{
	// 	C.Elem(1), C.Elem(1), C.Elem(1), C.Elem(1), C.Elem(1), C.Elem(1),
	// 	C.Elem(1), C.Elem(1), C.Elem(1), C.Elem(1), C.Elem(1), C.Elem(1),
	// 	C.Elem(1), C.Elem(1), C.Elem(1), C.Elem(1), C.Elem(1), C.Elem(1),
	// 	C.Elem(1), C.Elem(1), C.Elem(1), C.Elem(1), C.Elem(1), C.Elem(1),
	// 	C.Elem(1), C.Elem(1), C.Elem(1), C.Elem(1), C.Elem(1), C.Elem(1),
	// 	C.Elem(1), C.Elem(1), C.Elem(1), C.Elem(1), C.Elem(1), C.Elem(1)}}
	// logger.Println("DB*myvector:", MyMatrixMul(DB.Data, DB.Info.Packing, mymatrix, uint64(2)))
	// logger.Println()
	// logger.Println()
	// logger.Println("DB*A*s:", MyMatrixMul(MyMatrixMul(DB.Data, DB.Info.Packing, shared.Data[0], uint64(1)), 1, secret, 1))
	// logger.Println()
	// logger.Println()
	// logger.Println("A*s Using MatrixMul():", MatrixMul(shared.Data[0], secret))
	// logger.Println()
	// logger.Println()
	// logger.Println("A*s Using MyMatrixMul():", MyMatrixMul(shared.Data[0], 1, secret, 1))

	// H1 := MyMatrixMul(DB.Data, DB.Info.Packing, shared.Data[0], uint64(1))
	// fmt.Println("A:", shared.Data[0].Data[0], shared.Data[0].Data[1], "\t row:", shared.Data[0].Rows, "\t col:", shared.Data[0].Cols)
	// fmt.Println("H == H1 \t ", H == H1)
	// fmt.Println("hint*s:", MatrixMul(H, secret))
	// fmt.Println("hint*s:", MatrixMul(H1, secret))
	// for i := 0; i < 3*1024; i++ {
	// 	if H.Data[i] != H1.Data[i] {
	// 		fmt.Println("i=", i, "\tH.Data[i]", H.Data[i], "\tH1.Data[i]", H1.Data[i])
	// 		panic("H != H1")
	// 	}
	// 	// fmt.Println("hint:", H.Data[i], H.Data[1], "\t row:", H.Rows, "\t col:", H.Cols)
	// 	// fmt.Println("hint1:", H1.Data[i], H1.Data[1], "\t row:", H1.Rows, "\t col:", H1.Cols)
	// }
	// fmt.Println("secret:", secret.Data[0], secret.Data[1], "\t row:", secret.Rows, "\t col:", secret.Cols)
	// fmt.Println("s:", secret)

	// ratio := p.P / 2
	// offset := uint64(0)
	// for j := uint64(0); j < p.M; j++ {
	// 	offset += ratio * query.Data[0].Get(j, 0)
	// }
	// offset %= (1 << p.Logq)
	// offset = (1 << p.Logq) - offset

	secret := client.Data[0]
	// fmt.Println("err:", err.Data)
	H := offline.Data[0]
	interm := MatrixMul(H, secret)
	// interm := MyMatrixMul(DB.Data, DB.Info.Packing, MatrixMul(shared.Data[0], secret), 1)
	ans := answer.Data[batch_index]
	ans.MatrixSub(interm)
	// fmt.Println("recover ans.MatrixSub(interm):", ans.Data)

	// fmt.Println("DB row:", DB.Data.Rows, "\tDB Col:", DB.Data.Cols)
	// ans.MatrixSub(MyMatrixMul(DB.Data, DB.Info.Packing, err, uint64(1)))
	// fmt.Println("recover ΔDBUi:", ans.Data)
	var vals []uint64
	for i := uint64(0); i < ans.Rows; i++ {
		noised := uint64(ans.Data[i])
		denoised := ((noised + p.Delta()/2) / p.Delta()) % p.P
		vals = append(vals, denoised)
	}

	// for i := uint64(0); i < ans.Rows; i++ {
	// 	noised := uint64(ans.Data[i])

	// 	// Delta := p.Delta()
	// 	// v := (x + Delta/2) / Delta
	// 	// return v % p.P
	// 	denoised := (noised / p.Delta()) % p.P
	// 	vals = append(vals, noised)
	// 	vals = append(vals, denoised)
	// }
	// fmt.Println("Delta:", p.Delta())
	// ans.MatrixAdd(interm)
	return vals

	// return ans
}

func (pi *SimplePIR) Reset(DB *Database, p Params) {
	// Uncompress the database, and map its entries to the range [-p/2, p/2].
	DB.Unsquish()
	DB.Data.Sub(p.P / 2)
}
