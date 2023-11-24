// client.go
package main

import (
	"github.com/mora-2/simplepir/pir"
)

type shared_data struct {
	Info             pir.DBinfo
	P                pir.Params
	Comp             pir.CompressedState
	Offline_download pir.Msg
}
