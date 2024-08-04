package main

import (
	"fmt"

	"github.com/nicolito128/goetf"
)

func main() {
	var out string

	if err := goetf.Unmarshal(secretPhrase, &out); err != nil {
		panic(err)
	}

	fmt.Println("Phrase:", out)
}

// already encoded ETF data
var secretPhrase = []byte{
	131, 107, 0, 87, 34, 82, 101, 97, 108, 32, 115, 116, 117, 112, 105,
	100, 105, 116, 121, 32, 98, 101, 97, 116, 115, 32, 97, 114, 116, 105,
	102, 105, 99, 105, 97, 108, 32, 105, 110, 116, 101, 108, 108, 105,
	103, 101, 110, 99, 101, 32, 101, 118, 101, 114, 121, 32, 116, 105,
	109, 101, 46, 34, 32, 45, 32, 84, 101, 114, 114, 121, 32, 80, 114, 97,
	116, 99, 104, 101, 116, 116, 44, 32, 72, 111, 103, 102, 97, 116, 104,
	101, 114,
}
