/*
GoETF is a module capable of encoding and decoding byte slices of the ETF type.

The external term format is mainly used in Erlang's distribution system.
Occasionally, it's necessary to encode and decode this particular binary format for communication
between different APIs.

Start by importing the module:

	import "github.com/nicolito128/goetf"

You can use the Marshal function to encode values:

	func main() {
		phrase := "Hello, world!"

		data, err := goetf.Marshal(phrase)
		if err != nil {
			panic(err)
		}

		fmt.Println("Encoded:", data)
	}

Or use Unmarshal to decode the value:

	func main() {
		data := []byte{...}

		var out string
		if err := goetf.Unmarshal(data, &out); err != nil {
			panic(err)
		}

		fmt.Println("Output:", out)
	}

Alternatively, you can use the goetf.NewEncoder or goetf.NewDecoder functions
to create your own decoder/encoder.
*/
package goetf
