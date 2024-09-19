package server

import (
	"encoding/json"
	"fmt"
)

type Log struct {
	Duration float64 `json:"duration"`
	Request  Request `json:"request"`
	Size     int     `json:"size"`
	Status   int     `json:"status"`
}

type Headers struct {
	UserAgent []string `json:"User-Agent"`
}

type Request struct {
	RemoteIP string  `json:"remote_ip"`
	Proto    string  `json:"proto"`
	Method   string  `json:"method"`
	Host     string  `json:"host"`
	URI      string  `json:"uri"`
	Headers  Headers `json:"headers"`
}

func ParseLog(b []byte) (Log, error) {
	var l Log
	if err := json.Unmarshal(b, &l); err != nil {
		return Log{}, fmt.Errorf("unmarshalling log: %w", err)
	}

	return l, nil
}
