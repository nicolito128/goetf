package main

import (
	"fmt"

	"github.com/nicolito128/goetf"
)

// example struct
type User struct {
	Name   string `etf:"name"`
	Age    uint8  `etf:"age"`
	Active bool   `etf:"active"`
}

func main() {
	exampleUser := User{Name: "John Dee", Age: 34, Active: true}

	data, err := goetf.Marshal(exampleUser)
	if err != nil {
		panic(err)
	}
	fmt.Println("Data:", data)

	// Unmarshal
	var out User
	if err := goetf.Unmarshal(data, &out); err != nil {
		panic(err)
	}

	fmt.Println("Out:", out)
}
