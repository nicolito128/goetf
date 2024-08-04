/*
GoETF is a module capable of encoding and decoding byte slices of the ETF type.

The external term format is mainly used in Erlang's distribution system.
Occasionally, it's necessary to encode and decode this particular binary format for communication
between different APIs. This format offers the advantage of being faster and more lightweight
compared to traditional JSON.

You can start by importing the module:

	import "github.com/nicolito128/goetf"

And you can use the goetf.Marshal functions to encode values or goetf.Unmarshal to decode them.

Alternatively, you can use the goetf.NewEncoder or goetf.NewDecoder functions
to create your own decoder/encoder.
*/
package goetf
