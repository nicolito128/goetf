package main

import (
	"fmt"
	"time"

	"github.com/nicolito128/goetf"
)

// example struct
type User struct {
	Name   string `etf:"name"`
	Age    uint8  `etf:"age"`
	Active bool   `etf:"active"`
	Profile
}

type Profile struct {
	Status    int           `etf:"status"`
	VIP       *bool         `etf:"vip"`
	LastLogin time.Duration `etf:"last_login"`
}

func main() {
	exampleUser := User{
		Name:    "John Dee",
		Age:     34,
		Active:  true,
		Profile: Profile{Status: 0, VIP: nil, LastLogin: time.Duration(24 * 60)},
	}

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
