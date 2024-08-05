package goetf_test

import (
	"maps"
	"slices"
	"testing"

	"github.com/nicolito128/goetf"
)

func TestDecodeSmallInteger(t *testing.T) {
	want := uint8(128)
	b, err := goetf.Marshal(want)
	if err != nil {
		t.Fatal("marshal error:", err)
	}

	var out uint8
	if err := goetf.Unmarshal(b, &out); err != nil {
		t.Fatal("unmarshal error:", err)
	}

	if want != out {
		t.Errorf("unmarshal error: want = %v got = %v", want, out)
	}
}

func TestDecodeInteger(t *testing.T) {
	var want int32 = 131073
	b, err := goetf.Marshal(want)
	if err != nil {
		t.Fatal("marshal error:", err)
	}

	var out int32
	if err := goetf.Unmarshal(b, &out); err != nil {
		t.Fatal("unmarshal error:", err)
	}

	if want != out {
		t.Errorf("want = %v, got = %v", want, out)
	}
}

func TestDecodeSmallBig(t *testing.T) {
	var want int64 = -11111111111
	b, err := goetf.Marshal(want)
	if err != nil {
		t.Fatal("marshal error:", err)
	}

	var out int64
	if err := goetf.Unmarshal(b, &out); err != nil {
		t.Fatal("unmarshal error:", err)
	}

	if want != out {
		t.Errorf("want = %v, got = %v", want, out)
	}
}

func TestDecodeBig(t *testing.T) {
	{
		var want int64 = 27182818284590

		b, err := goetf.Marshal(want)
		if err != nil {
			t.Fatal("marshal error:", err)
		}

		var out int64
		if err := goetf.Unmarshal(b, &out); err != nil {
			t.Fatal("unmarshal error:", err)
		}

		if want != out {
			t.Errorf("want = %v, got = %v", want, out)
		}
	}
	/*{
		want := big.NewInt(314159265359)

		b, err := goetf.Marshal(want)
		if err != nil {
			t.Fatal("marshal error:", err)
		}

		out := big.NewInt(0)
		if err := goetf.Unmarshal(b, out); err != nil {
			t.Fatal("unmarshal error:", err)
		}

		if want.Cmp(out) != 0 {
			t.Errorf("want = %v, got = %v", want, out)
		}
	}*/
}

func TestDecodeSmallAtom(t *testing.T) {
	{
		want := true
		b, err := goetf.Marshal(want)
		if err != nil {
			t.Fatal("marshal error:", err)
		}

		var out bool
		if err := goetf.Unmarshal(b, &out); err != nil {
			t.Fatal("unmarshal error:", err)
		}

		if want != out {
			t.Errorf("unmarshal error: want = %v got = %v", want, out)
		}
	}
	{
		want := "ok"
		b, err := goetf.Marshal(want)
		if err != nil {
			t.Fatal("marshal error:", err)
		}

		var out string
		if err := goetf.Unmarshal(b, &out); err != nil {
			t.Fatal("unmarshal error:", err)
		}

		if want != out {
			t.Errorf("unmarshal error: want = %v got = %v", want, out)
		}
	}
}

func TestDecodeString(t *testing.T) {
	want := "hello, world"
	b, err := goetf.Marshal(want)
	if err != nil {
		t.Fatal("marshal error:", err)
	}

	var out string
	if err := goetf.Unmarshal(b, &out); err != nil {
		t.Fatal("unmarshal error:", err)
	}

	if want != out {
		t.Errorf("unmarshal error: want = %v got = %v", want, out)
	}
}

func TestDecodeBinary(t *testing.T) {
	{
		want := []byte{101, 111, 112, 107}
		b, err := goetf.Marshal(want)
		if err != nil {
			t.Fatal("marshal error:", err)
		}

		out := make([]byte, 0)
		if err := goetf.Unmarshal(b, &out); err != nil {
			t.Fatal("unmarshal error:", err)
		}

		if n := slices.Compare(want, out); n != 0 {
			t.Errorf("unmarshal error: want = %v got = %v", want, out)
		}
	}
	{
		want := []byte{91, 34, 116, 101, 115, 116, 34, 93, 44, 123, 34, 116, 34, 125}
		b, err := goetf.Marshal(want)
		if err != nil {
			t.Fatal("marshal error:", err)
		}

		out := make([]byte, 0)
		if err := goetf.Unmarshal(b, &out); err != nil {
			t.Fatal("unmarshal error:", err)
		}

		if n := slices.Compare(want, out); n != 0 {
			t.Errorf("unmarshal error: want = %v got = %v", want, out)
		}
	}
}

func TestDecodeBitBinary(t *testing.T) {
	want := []byte{128, 100, 99}
	b, err := goetf.Marshal(want)
	if err != nil {
		t.Fatal("marshal error:", err)
	}

	out := make([]byte, 0)
	if err := goetf.Unmarshal(b, &out); err != nil {
		t.Fatal("unmarshal error:", err)
	}

	if n := slices.Compare(want, out); n != 0 {
		t.Errorf("unmarshal error: want = %v got = %v", want, out)
	}
}

func TestDecodeTuples(t *testing.T) {
	{ // small tuple
		want := []int{1, 2, 3, 4, 5}

		b, err := goetf.Marshal(want)
		if err != nil {
			t.Fatal("marshal error:", err)
		}

		out := make([]int, 0)
		if err := goetf.Unmarshal(b, &out); err != nil {
			t.Fatal("unmarshal error:", err)
		}

		if !slices.Equal(want, out) {
			t.Errorf("unmarshal error: want = %v got = %v", want, out)
		}
	}
	{ // large tuple
		want := []int{110, 94, 257}

		b, err := goetf.Marshal(want)
		if err != nil {
			t.Fatal("marshal error:", err)
		}

		out := make([]int, 0)
		if err := goetf.Unmarshal(b, &out); err != nil {
			t.Fatal("unmarshal error:", err)
		}

		if !slices.Equal(want, out) {
			t.Errorf("unmarshal error: want = %v got = %v", want, out)
		}
	}
	{ // any tuple
		want := []any{1.0, 999, "two"}

		b, err := goetf.Marshal(want)
		if err != nil {
			t.Fatal("marshal error:", err)
		}

		out := make([]any, 0)
		if err := goetf.Unmarshal(b, &out); err != nil {
			t.Fatal("unmarshal error:", err)
		}

		if len(want) != len(out) {
			t.Errorf("unmarshal error: want = %v got = %v", want, out)
		}

		w1 := (want[0]).(float64)
		w3 := (want[2]).(string)
		o1 := (out[0]).(float64)
		o3 := (out[2]).(string)
		if w1 != o1 || w3 != o3 {
			t.Errorf("unmarshal error: want = %v got = %v", want, out)
		}
	}
	{ // tuple of pointers
		want := []*string{nil, nil, nil, nil, nil}

		b, err := goetf.Marshal(want)
		if err != nil {
			t.Fatal("marshal error:", err)
		}

		out := make([]*string, 5)
		if err := goetf.Unmarshal(b, &out); err != nil {
			t.Fatal("unmarshal error:", err)
		}

		if !slices.Equal(want, out) {
			t.Errorf("unmarshal error: want = %v got = %v", want, out)
		}
	}
	{ // just a nil
		want := (*[]byte)(nil)

		b, err := goetf.Marshal(want)
		if err != nil {
			t.Fatal("marshal error:", err)
		}

		var out []byte
		if err := goetf.Unmarshal(b, &out); err != nil {
			t.Fatal("unmarshal error:", err)
		}

		// asserting impossible condition
		if out != nil {
			t.Errorf("unmarshal error: want = %v got = %v", want, out)
		}
	}
}

func TestDecodeList(t *testing.T) {
	{
		want := [3]int{101, 201, 255}
		b, err := goetf.Marshal(want)
		if err != nil {
			t.Fatal("marshal error:", err)
		}

		var out [3]int
		if err := goetf.Unmarshal(b, &out); err != nil {
			t.Fatal("unmarshal error:", err)
		}

		if !slices.Equal(want[:], out[:]) {
			t.Errorf("unmarshal error: want = %v got = %v", want, out)
		}
	}
	{
		want := [2]float64{3.14159, 2.71828}
		b, err := goetf.Marshal(want)
		if err != nil {
			t.Fatal("marshal error:", err)
		}

		var out [2]float64
		if err := goetf.Unmarshal(b, &out); err != nil {
			t.Fatal("unmarshal error:", err)
		}

		if !slices.Equal(want[:], out[:]) {
			t.Errorf("unmarshal error: want = %v got = %v", want, out)
		}
	}
	{ // list of pointers
		want := [7]*float64{}
		b, err := goetf.Marshal(want)
		if err != nil {
			t.Fatal("marshal error:", err)
		}

		var out [7]*float64
		if err := goetf.Unmarshal(b, &out); err != nil {
			t.Fatal("unmarshal error:", err)
		}

		if !slices.Equal(want[:], out[:]) {
			t.Errorf("unmarshal error: want = %v got = %v", want, out)
		}
	}
	{ // pointer to list
		var want *[2]int = &[2]int{3, 5}

		b, err := goetf.Marshal(want)
		if err != nil {
			t.Fatal("marshal error:", err)
		}

		out := &[2]int{0, 0}
		if err := goetf.Unmarshal(b, out); err != nil {
			t.Fatal("unmarshal error:", err)
		}

		if want[0] != out[0] || want[1] != out[1] {
			t.Errorf("unmarshal error: want = %v got = %v", want, out)
		}
	}
}

func TestDecodeMap(t *testing.T) {
	{
		want := map[string]string{"name": "John", "position": "sysadmin"}
		b, err := goetf.Marshal(want)
		if err != nil {
			t.Fatal("marshal error:", err)
		}

		out := map[string]string{}
		if err := goetf.Unmarshal(b, out); err != nil {
			t.Fatal("unmarshal error:", err)
		}

		if !maps.Equal(want, out) {
			t.Errorf("unmarshal error: want = %v got = %v", want, out)
		}
	}
	{
		want := map[string]any{"a": true, "b": -3.14}
		b, err := goetf.Marshal(want)
		if err != nil {
			t.Fatal("marshal error:", err)
		}

		out := map[string]any{}
		if err := goetf.Unmarshal(b, out); err != nil {
			t.Fatal("unmarshal error:", err)
		}

		outa := out["a"].(bool)
		outb := out["b"].(float64)
		if want["a"] != outa || want["b"] != outb {
			t.Errorf("unmarshal error: want = %v got = %v", want, out)
		}
	}
	/*{
		type foo struct {
			Bar string `etf:"bar"`
		}

		want := map[string]*foo{"buz": {Bar: "buzbar"}, "baz": {Bar: "bazbar"}}
		b, err := goetf.Marshal(want)
		if err != nil {
			t.Fatal("marshal error:", err)
		}

		out := map[string]*foo{}
		if err := goetf.Unmarshal(b, out); err != nil {
			t.Fatal("unmarshal error:", err)
		}

		wBuz, oBuz := want["buz"], out["buz"]
		wBaz, oBaz := want["baz"], out["baz"]
		if wBuz.Bar != oBuz.Bar || wBaz.Bar != oBaz.Bar {
			t.Errorf("unmarshal error: want = %v got = %v", want, out)
		}
	}*/
}

func TestDecodeStruct(t *testing.T) {
	{
		type user struct {
			Name string `etf:"name"`
			Age  uint8  `etf:"age"`
		}

		want := user{"Mile", 22}
		b, err := goetf.Marshal(want)
		if err != nil {
			t.Fatal(err)
		}

		var out user
		if err := goetf.Unmarshal(b, &out); err != nil {
			t.Fatal(err)
		}

		if want.Name != out.Name || want.Age != out.Age {
			t.Errorf("unmarshal error: want = %v got = %v", want, out)
		}
	}
	{
		type c struct {
			B string `etf:"b"`
			A []int  `etf:"a"`
		}

		want := c{A: []int{1, 2}, B: "b"}
		b, err := goetf.Marshal(want)
		if err != nil {
			t.Fatal(err)
		}

		var out c
		if err := goetf.Unmarshal(b, &out); err != nil {
			t.Fatal(err)
		}

		if !slices.Equal(want.A, out.A) || want.B != out.B {
			t.Errorf("unmarshal error: want = %v got = %v", want, out)
		}
	}
	{
		type axes struct {
			X float64 `etf:"x"`
			Y float64 `etf:"y"`
		}

		want := []*axes{{X: 2.7, Y: 9.9}, {X: 1.23, Y: 2.01}}
		b, err := goetf.Marshal(want)
		if err != nil {
			t.Fatal(err)
		}

		out := make([]*axes, 0)
		if err := goetf.Unmarshal(b, &out); err != nil {
			t.Fatal(err)
		}

		if len(want) != len(out) {
			t.Errorf("unmarshal error: want = %v got = %v", want, out)
		} else {
			if (*want[0]).X != (*out[0]).X || (*want[0]).Y != (*out[0]).Y {
				t.Errorf("unmarshal error: want = %v got = %v", want, out)
			}

			if (*want[1]).X != (*out[1]).X || (*want[1]).Y != (*out[1]).Y {
				t.Errorf("unmarshal error: want = %v got = %v", want, out)
			}
		}
	}
	{
		type Activity struct {
			Online bool    `etf:"online"`
			Wallet float64 `etf:"wallet"`
		}

		type Client struct {
			Status *Activity `etf:"status"`
		}

		want := Client{Status: &Activity{true, 554.23}}
		b, err := goetf.Marshal(want)
		if err != nil {
			t.Fatal(err)
		}

		var out Client
		if err := goetf.Unmarshal(b, &out); err != nil {
			t.Fatal(err)
		}

		if want.Status.Online != out.Status.Online || want.Status.Wallet != out.Status.Wallet {
			t.Errorf("unmarshal error: want = %v got = %v", want, out)
		}
	}
	{
		type conn struct {
			Shards *[2]int `etf:"shards"`
			//Token  *string `etf:"token"`
		}

		//token := "12345"
		want := conn{&[2]int{1, 2}}
		b, err := goetf.Marshal(want)
		if err != nil {
			t.Fatal(err)
		}

		var out conn
		if err := goetf.Unmarshal(b, &out); err != nil {
			t.Fatal(err)
		}

		if want.Shards[0] != out.Shards[0] || want.Shards[1] != out.Shards[1] {
			t.Errorf("unmarshal error: want = %v got = %v", want, out)
		}
	}
}
