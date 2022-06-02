package model

import (
	"encoding/json"
	"time"
)

type Moc struct {
	Rest Rest `json:"rest"`
}

type Rest struct {
	Request  Request    `json:"request"`
	Response Response   `json:"response"`
	Callback []Callback `json:"callback"`
}

type Request struct {
	Method   string             `json:"method"`
	Body     json.RawMessage              `json:"body"`
	Endpoint string             `json:"endpoint"`
	Params   *map[string]string `json:"queryparams"`
	Headers  *map[string]string `json:"headers"`
}

type Response struct {
	Status  int                `json:"status"`
	Body    json.RawMessage              `json:"body"`
	Headers *map[string]string `json:"headers"`
	Delay   ProcessDelay       `json:"after"`
}

type Callback struct {
	CallbackRest CallbackRest `json:"rest"`
}

type CallbackRest struct {
	Request          Request      `json:"request"`
	Delay            ProcessDelay `json:"after"`
	AuthorizationKey string       `json:"token-generator"`
}

type ProcessDelay struct {
	delay int64
}

func (d *ProcessDelay) Delay() time.Duration {
	return time.Duration(d.delay)
}


func (d *ProcessDelay) UnmarshalJSON(data []byte) error {
	var input string
	if err := json.Unmarshal(data, &input); err != nil {
		return err
	}
	delayed, err := time.ParseDuration(input)
	if err != nil {
		return err
	}
	d.delay = int64(delayed)
	return nil
}
