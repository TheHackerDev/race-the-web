package toml

import (
	"bytes"
	"reflect"
	"testing"
	"time"

	"github.com/kylelemons/godebug/diff"
	"github.com/kylelemons/godebug/pretty"
)

func init() {
	// These settings make 'mismatch' output look nicer.
	pretty.DefaultConfig.ShortList = 80
	pretty.CompareConfig.ShortList = 80
}

type testMarshaler struct{ S string }

type testMarshalerPtr struct{ S string }

func (t testMarshaler) MarshalTOML() ([]byte, error) {
	return []byte(t.S), nil
}

func (t *testMarshalerPtr) MarshalTOML() ([]byte, error) {
	return []byte(t.S), nil
}

type testMarshalerRec struct{ replacement interface{} }

type testMarshalerRecPtr struct{ replacement interface{} }

func (t testMarshalerRec) MarshalTOML() (interface{}, error) {
	return t.replacement, nil
}

func (t *testMarshalerRecPtr) MarshalTOML() (interface{}, error) {
	return t.replacement, nil
}

var marshalTests = []struct {
	v      interface{}
	expect []byte
}{
	// single string:
	{
		v:      struct{ Name string }{"alice"},
		expect: []byte("name = \"alice\"\n"),
	},
	// single int:
	{
		v:      struct{ Age int }{7},
		expect: []byte("age = 7\n"),
	},
	// multiple fields:
	{
		v: struct {
			Name string
			Age  int
		}{"alice", 7},
		expect: []byte("name = \"alice\"\nage = 7\n"),
	},
	// ignored fields:
	{
		v: struct {
			Name string `toml:"-"`
			Age  int
		}{"alice", 7},
		expect: []byte("age = 7\n"),
	},
	// field name override:
	{
		v: struct {
			Name string `toml:"my_name"`
		}{"bob"},
		expect: []byte("my_name = \"bob\"\n"),
	},
	{
		v: struct {
			Name string `toml:"my_name,omitempty"`
		}{"bob"},
		expect: []byte("my_name = \"bob\"\n"),
	},
	// omitempty:
	{
		v: struct {
			Name string `toml:",omitempty"`
		}{"bob"},
		expect: []byte("name = \"bob\"\n"),
	},
	// slices:
	{
		v: struct {
			Ints []int
		}{[]int{1, 2, 3}},
		expect: []byte("ints = [1, 2, 3]\n"),
	},
	{
		v: struct {
			IntsOfInts [][]int
		}{[][]int{{}, {1}, {}, {2}, {3, 4}}},
		expect: []byte("ints_of_ints = [[], [1], [], [2], [3, 4]]\n"),
	},
	{
		v: struct {
			IntsOfInts [][]int
		}{[][]int{{1, 2}, {3, 4}}},
		expect: []byte("ints_of_ints = [[1, 2], [3, 4]]\n"),
	},
	// pointer:
	{
		v:      struct{ Named *Name }{&Name{First: "name"}},
		expect: []byte("[named]\nfirst = \"name\"\nlast = \"\"\n"),
	},
	// canon test document:
	{
		v:      theTestStruct(),
		expect: loadTestData("marshal-teststruct.toml"),
	},
	// funky map key types:
	{
		v: map[string]interface{}{
			"intKeys":       map[int]int{1: 1, 2: 2, 3: 3},
			"marshalerKeys": map[time.Time]int{time.Time{}: 1},
		},
		expect: loadTestData("marshal-funkymapkeys.toml"),
	},
	// Marshaler:
	{
		v: map[string]interface{}{
			"m1": testMarshaler{"1"},
			"m2": &testMarshaler{"2"},
			"m3": &testMarshalerPtr{"3"},
		},
		expect: loadTestData("marshal-marshaler.toml"),
	},
	// MarshalerRec:
	{
		v: map[string]interface{}{
			"m1": testMarshalerRec{1},
			"m2": &testMarshalerRec{2},
			"m3": &testMarshalerRecPtr{3},
			"sub": &testMarshalerRec{map[string]interface{}{
				"key": 1,
			}},
		},
		expect: loadTestData("marshal-marshalerrec.toml"),
	},
	// key escaping:
	{
		v: map[string]interface{}{
			"":    "empty",
			" ":   "space",
			"ʎǝʞ": "reverse",
			"1":   "number (not quoted)",
			"-":   "dash (not quoted)",
			"subtable with space": map[string]interface{}{
				"depth": 1,
				"subsubtable with space": map[string]interface{}{
					"depth": 2,
				},
			},
		},
		expect: loadTestData("marshal-key-escape.toml"),
	},
}

func TestMarshal(t *testing.T) {
	for _, test := range marshalTests {
		b, err := Marshal(test.v)
		if err != nil {
			t.Errorf("Unexpected error %v\nfor value:\n%s", err, pretty.Sprint(test.v))
		}
		if d := checkOutput(b, test.expect); d != "" {
			t.Errorf("Output mismatch:\nValue:\n%s\nDiff:\n%s", pretty.Sprint(test.v), d)
		}
	}
}

func TestMarshalRoundTrip(t *testing.T) {
	v := theTestStruct()
	b, err := Marshal(v)
	if err != nil {
		t.Error("Unexpected Marshal error:", err)
	}
	dest := testStruct{}
	if err := Unmarshal(b, &dest); err != nil {
		t.Error("Unmarshal error:", err)
	}
	if !reflect.DeepEqual(theTestStruct(), &dest) {
		t.Errorf("Unmarshaled value mismatch:\n%s", pretty.Compare(v, dest))
	}
}

func TestMarshalArrayTableEmptyParent(t *testing.T) {
	type Baz struct {
		Key int
	}
	type Bar struct {
		Baz Baz
	}
	type Foo struct {
		Bars []Bar
	}

	v := Foo{[]Bar{{Baz{1}}, {Baz{2}}}}
	b, err := Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	if d := checkOutput(b, loadTestData("marshal-arraytable-empty.toml")); d != "" {
		t.Errorf("Output mismatch:\n%s", d)
	}
}

func TestMarshalPointerError(t *testing.T) {
	type X struct{ Sub *X }
	want := &marshalNilError{reflect.TypeOf((*X)(nil))}

	if _, err := Marshal((*X)(nil)); !reflect.DeepEqual(err, want) {
		t.Errorf("Got %q, expected %q", err, want)
	}
	if _, err := Marshal(&X{nil}); !reflect.DeepEqual(err, want) {
		t.Errorf("Got %q, expected %q", err, want)
	}
}

func TestMarshalNonStruct(t *testing.T) {
	val := []string{}
	want := &marshalTableError{reflect.TypeOf(val)}
	if _, err := Marshal(val); !reflect.DeepEqual(err, want) {
		t.Errorf("Got %q, expected %q", err, want)
	}
}

func TestMarshalOmitempty(t *testing.T) {
	var x struct {
		ZeroArray  [0]int  `toml:",omitempty"`
		ZeroArray2 [10]int `toml:",omitempty"`
		Slice      []int   `toml:",omitempty"`
		Pointer    *int    `toml:",omitempty"`
		Int        int     `toml:",omitempty"`
	}
	out, err := Marshal(x)
	if err != nil {
		t.Fatal(err)
	}
	if len(out) > 0 {
		t.Fatalf("want no output, got %q", out)
	}
}

func checkOutput(got, want []byte) string {
	if bytes.Equal(got, want) {
		return ""
	}
	return diff.Diff(string(got), string(want))
}
