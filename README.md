# GoETF (Work-In-Progress)

> [!WARNING]
> The module is not yet at version `1.0.0` so it's not possible to ensure that there will be no breaking changes.

Go encoding module for ETF (External Term Format).

## Why GoETF?
The external term format is primarily used in Erlang's distribution mechanism. Sometimes, it's necessary to work with the encoding and decoding of this specific binary format, which is used for communication between different APIs. This format has the advantage of being faster and lighter than conventional JSON.

## Getting started

### Requirements

* `go v1.22+`

### Installation

    go get github.com/nicolito128/goetf

### Encoding example

```go
package main

import (
	"fmt"

	"github.com/nicolito128/goetf"
)

func main() {
	phrase := "Hello, world!"

	data, err := goetf.Marshal(phrase)
	if err != nil {
		panic(err)
	}

	fmt.Println("Encoded:", data)
}
```

### Decoding example

```go
package main

import (
	"fmt"

	"github.com/nicolito128/goetf"
)

func main() {
	var out int

    bin := []byte{131, 98, 0, 0, 1, 1}
	if err := goetf.Unmarshal(bin, &out); err != nil {
		panic(err)
	}

	fmt.Println("Out:", out)
}
```

## References

* [Erlang Runtime System Application (ERTS) - External Term Format](https://www.erlang.org/doc/apps/erts/erl_ext_dist.html)