package master

import "encoding/json"

// Msg is a message from/to a slave/master
type Msg struct {
	PaddingBefore string `json:"pab"`
	CmdType       string `json:"type"`
	Data          string `json:"data"`
	PaddingAfter  string `json:"pada"`
}

func (m *Msg) String() string {
	bytes, err := json.Marshal(&m)
	if err != nil {
		return ""
	}
	return string(bytes)
}
