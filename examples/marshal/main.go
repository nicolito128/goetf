package main

import (
	"fmt"

	"github.com/nicolito128/goetf"
)

func main() {
	phrase := "\"Real stupidity beats artificial intelligence every time.\" - Terry Pratchett, Hogfather"
	println(len(phrase))

	data, err := goetf.Marshal(phrase)
	if err != nil {
		panic(err)
	}

	fmt.Println("Phrase encoded:", data)
}
