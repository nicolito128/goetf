package main

import (
	"bytes"
	"fmt"

	"github.com/nicolito128/goetf"
)

var data = []byte{131, 116, 0, 0, 0, 4, 100, 0, 1, 100, 116, 0, 0, 0, 2, 100, 0, 6, 95, 116, 114, 97, 99, 101, 108, 0, 0, 0, 1, 109, 0, 0, 0, 46, 91, 34, 103, 97, 116, 101, 119, 97, 121, 45, 112, 114, 100, 45, 117, 115, 45, 101, 97, 115, 116, 49, 45, 100, 45, 52, 54, 106, 54, 34, 44, 123, 34, 109, 105, 99, 114, 111, 115, 34, 58, 48, 46, 48, 125, 93, 106, 100, 0, 18, 104, 101, 97, 114, 116, 98, 101, 97, 116, 95, 105, 110, 116, 101, 114, 118, 97, 108, 98, 0, 0, 161, 34, 100, 0, 2, 111, 112, 97, 10, 100, 0, 1, 115, 100, 0, 3, 110, 105, 108, 100, 0, 1, 116, 100, 0, 3, 110, 105, 108}

type OutStruct struct {
	Op int `etf:"op"`
	D  struct {
		HeartbeatInterval int `etf:"heartbeat_interval"`
	} `etf:"d"`
}

func main() {
	var res OutStruct

	d := goetf.NewDecoder(bytes.NewReader(data))
	if err := d.Decode(&res); err != nil {
		panic(err)
	}

	fmt.Println(res)
}