// config.go
package config

import (
	"github.com/mora-2/simplepir/pir"
)

type Shared_data struct {
	Info             pir.DBinfo
	P                pir.Params
	Comp             pir.CompressedState // A seed
	Offline_download pir.Msg             // hint H
}

type Pre_computed_data struct {
	DB           *pir.Database // unpacked DB
	Server_state pir.State
	Shared_state pir.State // A
	Pir_server   pir.SimplePIR
	P            pir.Params
}
