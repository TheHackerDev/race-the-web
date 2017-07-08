package toml

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/kylelemons/godebug/pretty"
)

func loadTestData(file string) []byte {
	f := filepath.Join("testdata", file)
	data, err := ioutil.ReadFile(f)
	if err != nil {
		panic(err)
	}
	return data
}

func mustTime(tm time.Time, err error) time.Time {
	if err != nil {
		panic(err)
	}
	return tm
}

type Name struct {
	First string
	Last  string
}
type Inline struct {
	Name  Name
	Point map[string]int
}
type Subtable struct {
	Key string
}
type Table struct {
	Key      string
	Subtable Subtable
	Inline   Inline
}
type W struct {
}
type Z struct {
	W W
}
type Y struct {
	Z Z
}
type X struct {
	Y Y
}
type Basic struct {
	Basic string
}
type Continued struct {
	Key1 string
	Key2 string
	Key3 string
}
type Multiline struct {
	Key1      string
	Key2      string
	Key3      string
	Continued Continued
}
type LiteralMultiline struct {
	Regex2 string
	Lines  string
}
type Literal struct {
	Winpath   string
	Winpath2  string
	Quoted    string
	Regex     string
	Multiline LiteralMultiline
}
type String struct {
	Basic     Basic
	Multiline Multiline
	Literal   Literal
}
type IntegerUnderscores struct {
	Key1 int
	Key2 int
	Key3 int
}
type Integer struct {
	Key1        int
	Key2        int
	Key3        int
	Key4        int
	Underscores IntegerUnderscores
}
type Fractional struct {
	Key1 float64
	Key2 float64
	Key3 float64
}
type Exponent struct {
	Key1 float64
	Key2 float64
	Key3 float64
}
type Both struct {
	Key float64
}
type FloatUnderscores struct {
	Key1 float64
	Key2 float64
}
type Float struct {
	Fractional  Fractional
	Exponent    Exponent
	Both        Both
	Underscores FloatUnderscores
}
type Boolean struct {
	True  bool
	False bool
}
type Datetime struct {
	Key1 time.Time
	Key2 time.Time
	Key3 time.Time
}
type Array struct {
	Key1 []int
	Key2 []string
	Key3 [][]int
	Key4 [][]interface{}
	Key5 []int
	Key6 []int
}
type Product struct {
	Name  string `toml:",omitempty"`
	Sku   int64  `toml:",omitempty"`
	Color string `toml:",omitempty"`
}
type Physical struct {
	Color string
	Shape string
}
type Variety struct {
	Name string
}
type Fruit struct {
	Name     string
	Physical Physical
	Variety  []Variety
}
type testStruct struct {
	Table    Table
	X        X
	String   String
	Integer  Integer
	Float    Float
	Boolean  Boolean
	Datetime Datetime
	Array    Array
	Products []Product
	Fruit    []Fruit
}

func theTestStruct() *testStruct {
	return &testStruct{
		Table: Table{
			Key: "value",
			Subtable: Subtable{
				Key: "another value",
			},
			Inline: Inline{
				Name: Name{
					First: "Tom",
					Last:  "Preston-Werner",
				},
				Point: map[string]int{
					"x": 1,
					"y": 2,
				},
			},
		},
		X: X{},
		String: String{
			Basic: Basic{
				Basic: "I'm a string. \"You can quote me\". Name\tJos\u00E9\nLocation\tSF.",
			},
			Multiline: Multiline{
				Key1: "One\nTwo",
				Key2: "One\nTwo",
				Key3: "One\nTwo",
				Continued: Continued{
					Key1: "The quick brown fox jumps over the lazy dog.",
					Key2: "The quick brown fox jumps over the lazy dog.",
					Key3: "The quick brown fox jumps over the lazy dog.",
				},
			},
			Literal: Literal{
				Winpath:  `C:\Users\nodejs\templates`,
				Winpath2: `\\ServerX\admin$\system32\`,
				Quoted:   `Tom "Dubs" Preston-Werner`,
				Regex:    `<\i\c*\s*>`,
				Multiline: LiteralMultiline{
					Regex2: `I [dw]on't need \d{2} apples`,
					Lines:  "The first newline is\ntrimmed in raw strings.\n   All other whitespace\n   is preserved.\n",
				},
			},
		},
		Integer: Integer{
			Key1: 99,
			Key2: 42,
			Key3: 0,
			Key4: -17,
			Underscores: IntegerUnderscores{
				Key1: 1000,
				Key2: 5349221,
				Key3: 12345,
			},
		},
		Float: Float{
			Fractional: Fractional{
				Key1: 1.0,
				Key2: 3.1415,
				Key3: -0.01,
			},
			Exponent: Exponent{
				Key1: 5e22,
				Key2: 1e6,
				Key3: -2e-2,
			},
			Both: Both{
				Key: 6.626e-34,
			},
			Underscores: FloatUnderscores{
				Key1: 9224617.445991228313,
				Key2: 1e100,
			},
		},
		Boolean: Boolean{
			True:  true,
			False: false,
		},
		Datetime: Datetime{
			Key1: mustTime(time.Parse(time.RFC3339Nano, "1979-05-27T07:32:00Z")),
			Key2: mustTime(time.Parse(time.RFC3339Nano, "1979-05-27T00:32:00-07:00")),
			Key3: mustTime(time.Parse(time.RFC3339Nano, "1979-05-27T00:32:00.999999-07:00")),
		},
		Array: Array{
			Key1: []int{1, 2, 3},
			Key2: []string{"red", "yellow", "green"},
			Key3: [][]int{{1, 2}, {3, 4, 5}},
			Key4: [][]interface{}{{int64(1), int64(2)}, {"a", "b", "c"}},
			Key5: []int{1, 2, 3},
			Key6: []int{1, 2},
		},
		Products: []Product{
			{Name: "Hammer", Sku: 738594937},
			{},
			{Name: "Nail", Sku: 284758393, Color: "gray"},
		},
		Fruit: []Fruit{
			{
				Name: "apple",
				Physical: Physical{
					Color: "red",
					Shape: "round",
				},
				Variety: []Variety{
					{Name: "red delicious"},
					{Name: "granny smith"},
				},
			},
			{
				Name: "banana",
				Variety: []Variety{
					{Name: "plantain"},
				},
			},
		},
	}
}

func TestUnmarshal(t *testing.T) {
	testUnmarshal(t, []testcase{
		{
			data:   string(loadTestData("test.toml")),
			expect: theTestStruct(),
		},
	})
}

type testcase struct {
	data   string
	err    error
	expect interface{}
}

func testUnmarshal(t *testing.T, testcases []testcase) {
	for _, test := range testcases {
		// Create a test value of the same type as expect.
		typ := reflect.TypeOf(test.expect)
		var val interface{}
		if typ.Kind() == reflect.Map {
			val = reflect.MakeMap(typ).Interface()
		} else if typ.Kind() == reflect.Ptr {
			val = reflect.New(typ.Elem()).Interface()
		} else {
			panic("invalid 'expect' type " + typ.String())
		}

		err := Unmarshal([]byte(test.data), val)
		if !reflect.DeepEqual(err, test.err) {
			t.Errorf("Error mismatch for input:\n%s\ngot:  %+v\nwant: %+v", test.data, err, test.err)
		}
		if err == nil && !reflect.DeepEqual(val, test.expect) {
			t.Errorf("Unmarshal value mismatch for input:\n%s\ndiff:\n%s", test.data, pretty.Compare(val, test.expect))
		}
	}
}

func TestUnmarshal_WithString(t *testing.T) {
	type testStruct struct {
		Str      string
		Key1     string
		Key2     string
		Key3     string
		Winpath  string
		Winpath2 string
		Quoted   string
		Regex    string
		Regex2   string
		Lines    string
	}
	testUnmarshal(t, []testcase{
		{
			data: `str = "I'm a string. \"You can quote me\". Name\tJos\u00E9\nLocation\tSF."`,
			expect: &testStruct{
				Str: "I'm a string. \"You can quote me\". Name\tJos\u00E9\nLocation\tSF.",
			},
		},
		{
			data:   string(loadTestData("unmarshal-string-1.toml")),
			expect: &testStruct{Key1: "One\nTwo", Key2: "One\nTwo", Key3: "One\nTwo"},
		},
		{
			data: string(loadTestData("unmarshal-string-2.toml")),
			expect: &testStruct{
				Key1: "The quick brown fox jumps over the lazy dog.",
				Key2: "The quick brown fox jumps over the lazy dog.",
				Key3: "The quick brown fox jumps over the lazy dog.",
			},
		},
		{
			data: string(loadTestData("unmarshal-string-3.toml")),
			expect: &testStruct{
				Winpath:  `C:\Users\nodejs\templates`,
				Winpath2: `\\ServerX\admin$\system32\`,
				Quoted:   `Tom "Dubs" Preston-Werner`,
				Regex:    `<\i\c*\s*>`,
			},
		},
		{
			data: string(loadTestData("unmarshal-string-4.toml")),
			expect: &testStruct{
				Regex2: `I [dw]on't need \d{2} apples`,
				Lines:  "The first newline is\ntrimmed in raw strings.\n   All other whitespace\n   is preserved.\n",
			},
		},
	})
}

func TestUnmarshal_WithInteger(t *testing.T) {
	type testStruct struct {
		Intval int64
	}
	testUnmarshal(t, []testcase{
		{`intval = 0`, nil, &testStruct{0}},
		{`intval = +0`, nil, &testStruct{0}},
		{`intval = -0`, nil, &testStruct{-0}},
		{`intval = 1`, nil, &testStruct{1}},
		{`intval = +1`, nil, &testStruct{1}},
		{`intval = -1`, nil, &testStruct{-1}},
		{`intval = 10`, nil, &testStruct{10}},
		{`intval = 777`, nil, &testStruct{777}},
		{`intval = 2147483647`, nil, &testStruct{2147483647}},
		{`intval = 2147483648`, nil, &testStruct{2147483648}},
		{`intval = +2147483648`, nil, &testStruct{2147483648}},
		{`intval = -2147483648`, nil, &testStruct{-2147483648}},
		{`intval = -2147483649`, nil, &testStruct{-2147483649}},
		{`intval = 9223372036854775807`, nil, &testStruct{9223372036854775807}},
		{`intval = +9223372036854775807`, nil, &testStruct{9223372036854775807}},
		{`intval = -9223372036854775808`, nil, &testStruct{-9223372036854775808}},
		{`intval = 1_000`, nil, &testStruct{1000}},
		{`intval = 5_349_221`, nil, &testStruct{5349221}},
		{`intval = 1_2_3_4_5`, nil, &testStruct{12345}},
		// overflow
		{
			data:   `intval = 9223372036854775808`,
			err:    lineErrorField(1, "toml.testStruct.Intval", &overflowError{reflect.Int64, "9223372036854775808"}),
			expect: &testStruct{},
		},
		{
			data:   `intval = +9223372036854775808`,
			err:    lineErrorField(1, "toml.testStruct.Intval", &overflowError{reflect.Int64, "+9223372036854775808"}),
			expect: &testStruct{},
		},
		{
			data:   `intval = -9223372036854775809`,
			err:    lineErrorField(1, "toml.testStruct.Intval", &overflowError{reflect.Int64, "-9223372036854775809"}),
			expect: &testStruct{},
		},
		// invalid _
		{`intval = _1_000`, lineError(1, errParse), &testStruct{}},
		{`intval = 1_000_`, lineError(1, errParse), &testStruct{}},
	})
}

func TestUnmarshal_WithFloat(t *testing.T) {
	type testStruct struct {
		Floatval float64
	}
	testUnmarshal(t, []testcase{
		{`floatval = 0.0`, nil, &testStruct{0.0}},
		{`floatval = +0.0`, nil, &testStruct{0.0}},
		{`floatval = -0.0`, nil, &testStruct{-0.0}},
		{`floatval = 0.1`, nil, &testStruct{0.1}},
		{`floatval = +0.1`, nil, &testStruct{0.1}},
		{`floatval = -0.1`, nil, &testStruct{-0.1}},
		{`floatval = 0.2`, nil, &testStruct{0.2}},
		{`floatval = +0.2`, nil, &testStruct{0.2}},
		{`floatval = -0.2`, nil, &testStruct{-0.2}},
		{`floatval = 1.0`, nil, &testStruct{1.0}},
		{`floatval = +1.0`, nil, &testStruct{1.0}},
		{`floatval = -1.0`, nil, &testStruct{-1.0}},
		{`floatval = 1.1`, nil, &testStruct{1.1}},
		{`floatval = +1.1`, nil, &testStruct{1.1}},
		{`floatval = -1.1`, nil, &testStruct{-1.1}},
		{`floatval = 3.1415`, nil, &testStruct{3.1415}},
		{`floatval = +3.1415`, nil, &testStruct{3.1415}},
		{`floatval = -3.1415`, nil, &testStruct{-3.1415}},
		{`floatval = 10.2e5`, nil, &testStruct{10.2e5}},
		{`floatval = +10.2e5`, nil, &testStruct{10.2e5}},
		{`floatval = -10.2e5`, nil, &testStruct{-10.2e5}},
		{`floatval = 10.2E5`, nil, &testStruct{10.2e5}},
		{`floatval = +10.2E5`, nil, &testStruct{10.2e5}},
		{`floatval = -10.2E5`, nil, &testStruct{-10.2e5}},
		{`floatval = 5e+22`, nil, &testStruct{5e+22}},
		{`floatval = 1e6`, nil, &testStruct{1e6}},
		{`floatval = -2E-2`, nil, &testStruct{-2E-2}},
		{`floatval = 6.626e-34`, nil, &testStruct{6.626e-34}},
		{`floatval = 9_224_617.445_991_228_313`, nil, &testStruct{9224617.445991228313}},
		{`floatval = 1e1_00`, nil, &testStruct{1e100}},
		{`floatval = 1e02`, nil, &testStruct{1e2}},
		{`floatval = _1e1_00`, lineError(1, errParse), &testStruct{}},
		{`floatval = 1e1_00_`, lineError(1, errParse), &testStruct{}},
	})
}

func TestUnmarshal_WithBoolean(t *testing.T) {
	type testStruct struct {
		Boolval bool
	}
	testUnmarshal(t, []testcase{
		{`boolval = true`, nil, &testStruct{true}},
		{`boolval = false`, nil, &testStruct{false}},
	})
}

func TestUnmarshal_WithDatetime(t *testing.T) {
	type testStruct struct {
		Datetimeval time.Time
	}
	testUnmarshal(t, []testcase{
		{`datetimeval = 1979-05-27T07:32:00Z`, nil, &testStruct{
			mustTime(time.Parse(time.RFC3339Nano, "1979-05-27T07:32:00Z")),
		}},
		{`datetimeval = 2014-09-13T12:37:39Z`, nil, &testStruct{
			mustTime(time.Parse(time.RFC3339Nano, "2014-09-13T12:37:39Z")),
		}},
		{`datetimeval = 1979-05-27T00:32:00-07:00`, nil, &testStruct{
			mustTime(time.Parse(time.RFC3339Nano, "1979-05-27T00:32:00-07:00")),
		}},
		{`datetimeval = 1979-05-27T00:32:00.999999-07:00`, nil, &testStruct{
			mustTime(time.Parse(time.RFC3339Nano, "1979-05-27T00:32:00.999999-07:00")),
		}},
		{`datetimeval = 1979-05-27`, nil, &testStruct{
			mustTime(time.Parse(time.RFC3339, "1979-05-27T00:00:00Z")),
		}},
		{`datetimeval = 07:32:00`, nil, &testStruct{
			mustTime(time.Parse(time.RFC3339, "0000-01-01T07:32:00Z")),
		}},
		{`datetimeval = 00:32:00.999999`, nil, &testStruct{
			mustTime(time.Parse(time.RFC3339Nano, "0000-01-01T00:32:00.999999Z")),
		}},
	})
}

func TestUnmarshal_WithArray(t *testing.T) {
	type arrays struct {
		Ints    []int
		Strings []string
	}

	testUnmarshal(t, []testcase{
		{`ints = []`, nil, &arrays{Ints: []int{}}},
		{`ints = [ 1 ]`, nil, &arrays{Ints: []int{1}}},
		{`ints = [ 1, 2, 3 ]`, nil, &arrays{Ints: []int{1, 2, 3}}},
		{`ints = [ 1, 2, 3, ]`, nil, &arrays{Ints: []int{1, 2, 3}}},
		{`strings = ["red", "yellow", "green"]`, nil, &arrays{Strings: []string{"red", "yellow", "green"}}},
		{
			data:   `strings = [ "all", 'strings', """are the same""", '''type''']`,
			expect: &arrays{Strings: []string{"all", "strings", "are the same", "type"}},
		},
		{`arrayval = [[1,2],[3,4,5]]`, nil, &struct{ Arrayval [][]int }{
			[][]int{
				{1, 2},
				{3, 4, 5},
			},
		}},
		{`arrayval = [ [ 1, 2 ], ["a", "b", "c"] ] # this is ok`, nil,
			&struct{ Arrayval [][]interface{} }{
				[][]interface{}{
					{int64(1), int64(2)},
					{"a", "b", "c"},
				},
			}},
		{`arrayval = [ [ 1, 2 ], [ [3, 4], [5, 6] ] ] # this is ok`, nil,
			&struct{ Arrayval [][]interface{} }{
				[][]interface{}{
					{int64(1), int64(2)},
					{
						[]interface{}{int64(3), int64(4)},
						[]interface{}{int64(5), int64(6)},
					},
				},
			}},
		{`arrayval = [ [ 1, 2 ], [ [3, 4], [5, 6], [7, 8] ] ] # this is ok`, nil,
			&struct{ Arrayval [][]interface{} }{
				[][]interface{}{
					{int64(1), int64(2)},
					{
						[]interface{}{int64(3), int64(4)},
						[]interface{}{int64(5), int64(6)},
						[]interface{}{int64(7), int64(8)},
					},
				},
			}},
		{`arrayval = [ [[ 1, 2 ]], [3, 4], [5, 6] ] # this is ok`, nil,
			&struct{ Arrayval [][]interface{} }{
				[][]interface{}{
					{
						[]interface{}{int64(1), int64(2)},
					},
					{int64(3), int64(4)},
					{int64(5), int64(6)},
				},
			}},
		{
			data:   `ints = [ 1, 2.0 ] # note: this is NOT ok`,
			err:    lineErrorField(1, "toml.arrays.Ints", errArrayMultiType),
			expect: &arrays{},
		},
		// whitespace + comments
		{string(loadTestData("unmarshal-array-1.toml")), nil, &arrays{Ints: []int{1, 2, 3}}},
		{string(loadTestData("unmarshal-array-2.toml")), nil, &arrays{Ints: []int{1, 2, 3}}},
		{string(loadTestData("unmarshal-array-3.toml")), nil, &arrays{Ints: []int{1, 2, 3}}},
		{string(loadTestData("unmarshal-array-4.toml")), nil, &arrays{Ints: []int{1, 2, 3}}},
		{string(loadTestData("unmarshal-array-5.toml")), nil, &arrays{Ints: []int{1, 2, 3}}},
		{string(loadTestData("unmarshal-array-6.toml")), nil, &arrays{Ints: []int{1, 2, 3}}},
		// parse errors
		{`ints = [ , ]`, lineError(1, errParse), &arrays{}},
		{`ints = [ , 1 ]`, lineError(1, errParse), &arrays{}},
		{`ints = [ 1 2 ]`, lineError(1, errParse), &arrays{}},
		{`ints = [ 1 , , 2 ]`, lineError(1, errParse), &arrays{}},
	})
}

func TestUnmarshal_WithTable(t *testing.T) {
	type W struct{}
	type Z struct {
		W W
	}
	type Y struct {
		Z Z
	}
	type X struct {
		Y Y
	}
	type A struct {
		D int
		B struct {
			C int
		}
	}
	type testStruct struct {
		Table struct {
			Key string
		}
		Dog struct {
			Tater struct{}
		}
		X X
		A A
	}
	type testIgnoredFieldStruct struct {
		Ignored string `toml:"-"`
	}
	type testNamedFieldStruct struct {
		Named string `toml:"Named_Field"`
	}
	type testQuotedKeyStruct struct {
		Dog struct {
			TaterMan struct {
				Type string
			} `toml:"tater.man"`
		}
	}
	type testQuotedKeyWithWhitespaceStruct struct {
		Dog struct {
			TaterMan struct {
				Type string
			} `toml:"tater . man"`
		}
	}
	type testStructWithMap struct {
		Servers map[string]struct {
			IP string
			DC string
		}
	}
	type withTableArray struct {
		Tabarray []map[string]string
	}

	testUnmarshal(t, []testcase{
		{`[table]`, nil, &testStruct{}},
		{`[table]
key = "value"`, nil,
			&testStruct{
				Table: struct {
					Key string
				}{
					Key: "value",
				},
			}},
		{`[dog.tater]`, nil,
			&testStruct{
				Dog: struct {
					Tater struct{}
				}{
					Tater: struct{}{},
				},
			}},
		{`[dog."tater.man"]
type = "pug"`, nil,
			&testQuotedKeyStruct{
				Dog: struct {
					TaterMan struct {
						Type string
					} `toml:"tater.man"`
				}{
					TaterMan: struct {
						Type string
					}{
						Type: "pug",
					},
				},
			}},
		{`[dog."tater . man"]
type = "pug"`, nil,
			&testQuotedKeyWithWhitespaceStruct{
				Dog: struct {
					TaterMan struct {
						Type string
					} `toml:"tater . man"`
				}{
					TaterMan: struct {
						Type string
					}{
						Type: "pug",
					},
				},
			}},
		{`[x.y.z.w] # for this to work`, nil,
			&testStruct{
				X: X{},
			}},
		{`[ x .  y  . z . w ]`, nil,
			&testStruct{
				X: X{},
			}},
		{`[ x . "y" . z . "w" ]`, nil,
			&testStruct{
				X: X{},
			}},
		{`table = {}`, nil, &testStruct{}},
		{`table = { key = "value" }`, nil, &testStruct{
			Table: struct {
				Key string
			}{
				Key: "value",
			},
		}},
		{`x = { y = { "z" = { w = {} } } }`, nil, &testStruct{X: X{}}},
		{`[a.b]
c = 1

[a]
d = 2`, nil,
			&testStruct{
				A: struct {
					D int
					B struct {
						C int
					}
				}{
					D: 2,
					B: struct {
						C int
					}{
						C: 1,
					},
				},
			}},
		{
			data:   `Named_Field = "value"`,
			expect: &testNamedFieldStruct{Named: "value"},
		},
		{
			data: string(loadTestData("unmarshal-table-withmap.toml")),
			expect: &testStructWithMap{
				Servers: map[string]struct {
					IP string
					DC string
				}{
					"alpha": {IP: "10.0.0.1", DC: "eqdc10"},
					"beta":  {IP: "10.0.0.2", DC: "eqdc10"},
				},
			},
		},
		{
			data: string(loadTestData("unmarshal-table-withinline.toml")),
			expect: map[string]withTableArray{
				"tab1": {Tabarray: []map[string]string{{"key": "1"}}},
				"tab2": {Tabarray: []map[string]string{{"key": "2"}}},
			},
		},

		// errors
		{
			data:   string(loadTestData("unmarshal-table-conflict-1.toml")),
			err:    lineError(7, fmt.Errorf("table `a' is in conflict with table in line 4")),
			expect: &testStruct{},
		},
		{
			data:   string(loadTestData("unmarshal-table-conflict-2.toml")),
			err:    lineError(7, fmt.Errorf("table `a.b' is in conflict with line 5")),
			expect: &testStruct{},
		},
		{
			data:   string(loadTestData("unmarshal-table-conflict-3.toml")),
			err:    lineError(8, fmt.Errorf("key `b' is in conflict with table in line 4")),
			expect: &testStruct{},
		},
		{`[]`, lineError(1, errParse), &testStruct{}},
		{`[a.]`, lineError(1, errParse), &testStruct{}},
		{`[a..b]`, lineError(1, errParse), &testStruct{}},
		{`[.b]`, lineError(1, errParse), &testStruct{}},
		{`[.]`, lineError(1, errParse), &testStruct{}},
		{` = "no key name" # not allowed`, lineError(1, errParse), &testStruct{}},
		{
			data:   `ignored = "value"`,
			err:    lineError(1, fmt.Errorf("field corresponding to `ignored' in toml.testIgnoredFieldStruct cannot be set through TOML")),
			expect: &testIgnoredFieldStruct{},
		},
		{
			data:   `"-" = "value"`,
			err:    lineError(1, fmt.Errorf("field corresponding to `-' is not defined in toml.testIgnoredFieldStruct")),
			expect: &testIgnoredFieldStruct{},
		},
		{
			data:   `named = "value"`,
			err:    lineError(1, fmt.Errorf("field corresponding to `named' is not defined in toml.testNamedFieldStruct")),
			expect: &testNamedFieldStruct{},
		},
		{
			data: `
[a]
d = 2
y = 3
`,
			err:    lineError(4, fmt.Errorf("field corresponding to `y' is not defined in toml.A")),
			expect: &testStruct{},
		},
	})
}

func TestUnmarshal_WithEmbeddedStruct(t *testing.T) {
	type TestEmbStructA struct {
		A string
	}
	testUnmarshal(t, []testcase{
		{
			data: `a = "value"`,
			expect: &struct {
				TestEmbStructA
				A string
			}{
				A: "value",
			},
		},
		{
			data: `a = "value"`,
			expect: &struct {
				A string
				TestEmbStructA
			}{
				A: "value",
			},
		},
	})
}

func TestUnmarshal_WithArrayTable(t *testing.T) {
	type Product struct {
		Name  string
		SKU   int64
		Color string
	}
	type Physical struct {
		Color string
		Shape string
	}
	type Variety struct {
		Name string
	}
	type Fruit struct {
		Name     string
		Physical Physical
		Variety  []Variety
	}
	type testStruct struct {
		Products []Product
		Fruit    []Fruit
	}
	type testStructWithMap struct {
		Fruit []map[string][]struct {
			Name string
		}
	}
	testUnmarshal(t, []testcase{
		{
			data: string(loadTestData("unmarshal-arraytable.toml")),
			expect: &testStruct{
				Products: []Product{
					{Name: "Hammer", SKU: 738594937},
					{},
					{Name: "Nail", SKU: 284758393, Color: "gray"},
				},
			},
		},
		{
			data: string(loadTestData("unmarshal-arraytable-inline.toml")),
			expect: &testStruct{
				Products: []Product{
					{Name: "Hammer", SKU: 738594937},
					{},
					{Name: "Nail", SKU: 284758393, Color: "gray"},
				},
			},
		},
		{
			data: string(loadTestData("unmarshal-arraytable-nested-1.toml")),
			expect: &testStruct{
				Fruit: []Fruit{
					{
						Name: "apple",
						Physical: Physical{
							Color: "red",
							Shape: "round",
						},
						Variety: []Variety{
							{Name: "red delicious"},
							{Name: "granny smith"},
						},
					},
					{
						Name: "banana",
						Physical: Physical{
							Color: "yellow",
							Shape: "lune",
						},
						Variety: []Variety{
							{Name: "plantain"},
						},
					},
				},
			},
		},
		{
			data: string(loadTestData("unmarshal-arraytable-nested-2.toml")),
			expect: &testStructWithMap{
				Fruit: []map[string][]struct {
					Name string
				}{
					{"variety": {{Name: "red delicious"}, {Name: "granny smith"}}},
					{"variety": {{Name: "plantain"}}, "area": {{Name: "phillippines"}}},
				},
			},
		},
		{
			data: string(loadTestData("unmarshal-arraytable-nested-3.toml")),
			expect: &testStructWithMap{
				Fruit: []map[string][]struct {
					Name string
				}{
					{"variety": {{Name: "red delicious"}, {Name: "granny smith"}}},
					{"variety": {{Name: "plantain"}}, "area": {{Name: "phillippines"}}},
				},
			},
		},

		// errors
		{
			data:   string(loadTestData("unmarshal-arraytable-conflict-1.toml")),
			err:    lineError(10, fmt.Errorf("table `fruit.variety' is in conflict with array table in line 6")),
			expect: &testStruct{},
		},
		{
			data:   string(loadTestData("unmarshal-arraytable-conflict-2.toml")),
			err:    lineError(10, fmt.Errorf("array table `fruit.variety' is in conflict with table in line 6")),
			expect: &testStruct{},
		},
		{
			data:   string(loadTestData("unmarshal-arraytable-conflict-3.toml")),
			err:    lineError(8, fmt.Errorf("array table `fruit.variety' is in conflict with table in line 5")),
			expect: &testStruct{},
		},
	})
}

type testUnmarshalerString string

func (u *testUnmarshalerString) UnmarshalTOML(data []byte) error {
	*u = testUnmarshalerString("Unmarshaled: " + string(data))
	return nil
}

type testUnmarshalerStruct struct {
	Title  string
	Author testUnmarshalerString
}

func (u *testUnmarshalerStruct) UnmarshalTOML(data []byte) error {
	u.Title = "Unmarshaled: " + string(data)
	return nil
}

func TestUnmarshal_WithUnmarshaler(t *testing.T) {
	type testStruct struct {
		Title         testUnmarshalerString
		MaxConn       testUnmarshalerString
		Ports         testUnmarshalerString
		Servers       testUnmarshalerString
		Table         testUnmarshalerString
		Arraytable    testUnmarshalerString
		ArrayOfStruct []testUnmarshalerStruct
	}
	data := loadTestData("unmarshal-unmarshaler.toml")
	var v testStruct
	if err := Unmarshal(data, &v); err != nil {
		t.Fatal(err)
	}
	actual := v
	expect := testStruct{
		Title:      `Unmarshaled: "testtitle"`,
		MaxConn:    `Unmarshaled: 777`,
		Ports:      `Unmarshaled: [8080, 8081, 8082]`,
		Servers:    `Unmarshaled: [1, 2, 3]`,
		Table:      "Unmarshaled: [table]\nname = \"alice\"",
		Arraytable: "Unmarshaled: [[arraytable]]\nname = \"alice\"\n[[arraytable]]\nname = \"bob\"",
		ArrayOfStruct: []testUnmarshalerStruct{
			{
				Title:  "Unmarshaled: [[array_of_struct]]\ntitle = \"Alice's Adventures in Wonderland\"\nauthor = \"Lewis Carroll\"",
				Author: "",
			},
		},
	}
	if !reflect.DeepEqual(actual, expect) {
		t.Errorf(`toml.Unmarshal(data, &v); v => %#v; want %#v`, actual, expect)
	}
}

func TestUnmarshal_WithUnmarshalerForTopLevelStruct(t *testing.T) {
	data := `title = "Alice's Adventures in Wonderland"
author = "Lewis Carroll"
`
	var v testUnmarshalerStruct
	if err := Unmarshal([]byte(data), &v); err != nil {
		t.Fatal(err)
	}
	actual := v
	expect := testUnmarshalerStruct{
		Title: `Unmarshaled: title = "Alice's Adventures in Wonderland"
author = "Lewis Carroll"
`,
		Author: "",
	}
	if !reflect.DeepEqual(actual, expect) {
		t.Errorf(`toml.Unmarshal(data, &v); v => %#v; want %#v`, actual, expect)
	}
}

type testTextUnmarshaler string

var errTextUnmarshaler = errors.New("UnmarshalText called with data = error")

func (x *testTextUnmarshaler) UnmarshalText(data []byte) error {
	*x = testTextUnmarshaler("Unmarshaled: " + string(data))
	if string(data) == "error" {
		return errTextUnmarshaler
	}
	return nil
}

func TestUnmarshal_WithTextUnmarshaler(t *testing.T) {
	type testStruct struct {
		Str        testTextUnmarshaler
		Int        testTextUnmarshaler
		Float      testTextUnmarshaler
		Arraytable []testStruct
	}

	tests := []testcase{
		{
			data: string(loadTestData("unmarshal-textunmarshaler.toml")),
			expect: &testStruct{
				Str:        "Unmarshaled: str",
				Int:        "Unmarshaled: 11",
				Float:      "Unmarshaled: 12.0",
				Arraytable: []testStruct{{Str: "Unmarshaled: str2", Int: "Unmarshaled: 22", Float: "Unmarshaled: 23.0"}},
			},
		},
		{
			data:   `str = "error"`,
			expect: &testStruct{Str: "Unmarshaled: error"},
			err:    lineErrorField(1, "toml.testStruct.Str", errTextUnmarshaler),
		},
	}
	testUnmarshal(t, tests)
}

type testUnmarshalerRecString string

func (u *testUnmarshalerRecString) UnmarshalTOML(fn func(interface{}) error) error {
	var s string
	if err := fn(&s); err != nil {
		return err
	}
	*u = testUnmarshalerRecString("Unmarshaled: " + s)
	return nil
}

type testUnmarshalerRecStruct struct {
	a, b int
}

func (u *testUnmarshalerRecStruct) UnmarshalTOML(fn func(interface{}) error) error {
	var uu struct{ A, B int }
	if err := fn(&uu); err != nil {
		return err
	}
	u.a, u.b = uu.A, uu.B
	return nil
}

func TestUnmarshal_WithUnmarshalerRec(t *testing.T) {
	type testStruct struct {
		String     testUnmarshalerRecString
		Struct     testUnmarshalerRecStruct
		Arraytable []testStruct
	}
	var v testStruct
	err := Unmarshal(loadTestData("unmarshal-unmarshalerrec.toml"), &v)
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}
	expect := testStruct{
		String: "Unmarshaled: str1",
		Struct: testUnmarshalerRecStruct{a: 1, b: 2},
		Arraytable: []testStruct{
			{
				String: "Unmarshaled: str2",
				Struct: testUnmarshalerRecStruct{a: 3, b: 4},
			},
		},
	}
	if !reflect.DeepEqual(v, expect) {
		t.Errorf(`toml.Unmarshal(data, &v); v => %#v; want %#v`, v, expect)
	}
}

func TestUnmarshal_WithMultibyteString(t *testing.T) {
	type testStruct struct {
		Name    string
		Numbers []string
	}
	v := testStruct{}
	data := `name = "七一〇七"
numbers = ["壱", "弐", "参"]
`
	if err := Unmarshal([]byte(data), &v); err != nil {
		t.Fatal(err)
	}
	actual := v
	expect := testStruct{
		Name:    "七一〇七",
		Numbers: []string{"壱", "弐", "参"},
	}
	if !reflect.DeepEqual(actual, expect) {
		t.Errorf(`toml.Unmarshal([]byte(data), &v); v => %#v; want %#v`, actual, expect)
	}
}

func TestUnmarshal_WithPointers(t *testing.T) {
	type Inline struct {
		Key1 string
		Key2 *string
		Key3 **string
	}
	type Table struct {
		Key1 *string
		Key2 **string
		Key3 ***string
	}
	type testStruct struct {
		Inline *Inline
		Tables []*Table
	}
	type testStruct2 struct {
		Inline **Inline
		Tables []**Table
	}
	type testStruct3 struct {
		Inline ***Inline
		Tables []***Table
	}
	s1 := "a"
	s2 := &s1
	s3 := &s2
	s4 := &s3
	s5 := "b"
	s6 := &s5
	s7 := &s6
	s8 := &s7
	i1 := &Inline{"test", s2, s7}
	i2 := &i1
	i3 := &i2
	t1 := &Table{s2, s3, s4}
	t2 := &Table{s6, s7, s8}
	t3 := &t1
	t4 := &t2
	sc := &testStruct{
		Inline: i1, Tables: []*Table{t1, t2},
	}
	data := string(loadTestData("unmarshal-pointer.toml"))
	testUnmarshal(t, []testcase{
		{data, nil, sc},
		{data, nil, &testStruct2{
			Inline: i2,
			Tables: []**Table{&t1, &t2},
		}},
		{data, nil, &testStruct3{
			Inline: i3,
			Tables: []***Table{&t3, &t4},
		}},
	})
}

// This test checks that maps can be unmarshaled into directly.
func TestUnmarshalMap(t *testing.T) {
	testUnmarshal(t, []testcase{
		{
			data: `
name = "evan"
foo = 1
`,
			expect: map[string]interface{}{"name": "evan", "foo": int64(1)},
		},
		{
			data: `[""]
a = 1
`,
			expect: map[string]interface{}{"": map[string]interface{}{"a": int64(1)}},
		},
		{
			data: `["table"]
a = 1
`,
			expect: map[string]interface{}{"table": map[string]interface{}{"a": int64(1)}},
		},
		{
			data: `["\u2222"]
		a = 1
		`,
			expect: map[string]interface{}{"\u2222": map[string]interface{}{"a": int64(1)}},
		},
		{
			data: `[p]
first = "evan"
`,
			expect: map[string]*Name{"p": {First: "evan"}},
		},
		{
			data: `foo = 1
bar = 2
`,
			expect: map[testTextUnmarshaler]int{"Unmarshaled: foo": 1, "Unmarshaled: bar": 2},
		},
		{
			data: `"foo" = 1
"foo.bar" = 2
`,
			expect: map[testTextUnmarshaler]int{"Unmarshaled: foo": 1, "Unmarshaled: foo.bar": 2},
		},

		{
			data: `1 = 1
-2 = 2
`,
			expect: map[int]int{1: 1, -2: 2},
		},
		{
			data: `1 = 1
-129 = 2
`,
			expect: map[int8]int{1: 1},
			err:    lineError(2, &overflowError{reflect.Int8, "-129"}),
		},
	})
}

func TestUnmarshal_WithQuotedKeyValue(t *testing.T) {
	type nestedStruct struct {
		Truthy bool
	}
	type testStruct struct {
		Table map[string]nestedStruct
	}

	testUnmarshal(t, []testcase{
		{data: `"a" = 1`, expect: map[string]int{"a": 1}},
		{data: `"a.b" = 1`, expect: map[string]int{"a.b": 1}},
		{data: `"\u2222" = 1`, expect: map[string]int{"\u2222": 1}},
		{data: `"\"" = 1`, expect: map[string]int{"\"": 1}},
		{data: `"" = 1`, expect: map[string]int{"": 1}},
		{data: `'a' = 1`, expect: map[string]int{}, err: lineError(1, errParse)},
		// Inline tables:
		{
			data: `
[table]
"some.key" = {truthy = true}
`,
			expect: &testStruct{Table: map[string]nestedStruct{
				"some.key": {Truthy: true},
			}},
		},
		{
			data: `
"some.key" = [{truthy = true}]
`,
			expect: map[string][]nestedStruct{
				"some.key": {{Truthy: true}},
			},
		},
	})
}

func TestUnmarshal_WithCustomPrimitiveType(t *testing.T) {
	type (
		String string
		Int    int
		Bool   bool
	)
	type X struct {
		S String
		I Int
		B Bool
	}

	input := `
s = "string"
i = 1
b = true
`
	testUnmarshal(t, []testcase{
		{input, nil, &X{"string", 1, true}},
	})
}

func TestUnmarshal_WithInterface(t *testing.T) {
	var exp interface{} = map[string]interface{}{
		"string":   "string",
		"int":      int64(3),
		"float":    float64(4),
		"boolean":  true,
		"datetime": mustTime(time.Parse(time.RFC3339Nano, "1979-05-27T00:32:00.999999-07:00")),
		"array":    []interface{}{int64(1), int64(2), int64(3)},
		"inline":   map[string]interface{}{"key": "value"},
		"table":    map[string]interface{}{"key": "value"},
		"arraytable": []interface{}{
			map[string]interface{}{"key": "value"},
			map[string]interface{}{"key": "value"},
		},
	}

	type nonemptyIf interface {
		method()
	}
	nonemptyIfType := reflect.TypeOf((*nonemptyIf)(nil)).Elem()

	data := string(loadTestData("unmarshal-interface.toml"))
	testUnmarshal(t, []testcase{
		{data, nil, &exp},
		// can't unmarshal into non-empty interface{}
		{`v = "string"`, lineError(1, &unmarshalTypeError{"string", "", nonemptyIfType}), map[string]nonemptyIf{}},
		{`v = 1`, lineError(1, &unmarshalTypeError{"integer", "", nonemptyIfType}), map[string]nonemptyIf{}},
		{`v = 1.0`, lineError(1, &unmarshalTypeError{"float", "", nonemptyIfType}), map[string]nonemptyIf{}},
		{`v = true`, lineError(1, &unmarshalTypeError{"boolean", "", nonemptyIfType}), map[string]nonemptyIf{}},
		{`v = [1, 2]`, lineError(1, &unmarshalTypeError{"array", "slice", nonemptyIfType}), map[string]nonemptyIf{}},
		{`[v]`, lineError(1, &unmarshalTypeError{"table", "struct or map", nonemptyIfType}), map[string]nonemptyIf{}},
		{`[[v]]`, lineError(1, &unmarshalTypeError{"array table", "slice", nonemptyIfType}), map[string]nonemptyIf{}},
	})
}
