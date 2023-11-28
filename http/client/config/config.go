// client.go
package config

import (
	"github.com/mora-2/simplepir/pir"
)

type Shared_data struct {
	Info             pir.DBinfo
	P                pir.Params
	Comp             pir.CompressedState
	Offline_download pir.Msg
}

type Offline_data struct {
	Info             pir.DBinfo
	P                pir.Params
	Shared_state     pir.State
	Offline_download pir.Msg
}

type IP_Conn struct {
	IpAddr      string
	OnlinePort  uint32
	OfflinePort uint32
}
