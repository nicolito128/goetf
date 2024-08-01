package main

import (
	"fmt"

	"github.com/nicolito128/goetf"
)

type User struct {
	Name   string `etf:"name"`
	Age    uint8  `etf:"age"`
	Active bool   `etf:"active"`
	Wallet int32  `etf:"wallet"`
}

var data = []byte{
	131,
	116, 0, 0, 0, 4,
	119, 4, 110, 97, 109, 101, 107, 0, 4, 74, 111, 104, 110,
	119, 3, 97, 103, 101, 97, 21,
	119, 6, 97, 99, 116, 105, 118, 101, 119, 4, 116, 114, 117, 101,
	119, 6, 119, 97, 108, 108, 101, 116, 98, 0, 0, 1, 244,
}

func main() {
	var u User

	d := goetf.NewDecoder(data)
	if err := d.Decode(&u); err != nil {
		panic(err)
	}

	fmt.Println(u)
}
