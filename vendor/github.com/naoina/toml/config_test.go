package toml

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
)

func TestConfigNormField(t *testing.T) {
	cfg := Config{NormFieldName: func(reflect.Type, string) string { return "a" }}

	var x struct{ B int }
	input := []byte(`a = 1`)
	if err := cfg.Unmarshal(input, &x); err != nil {
		t.Fatal(err)
	}
	if x.B != 1 {
		t.Fatalf("wrong value after Unmarshal: got %d, want %d", x.B, 1)
	}

	dec := cfg.NewDecoder(strings.NewReader(`a = 2`))
	if err := dec.Decode(&x); err != nil {
		t.Fatal(err)
	}
	if x.B != 2 {
		t.Fatalf("wrong value after Decode: got %d, want %d", x.B, 2)
	}

	tbl, _ := Parse([]byte(`a = 3`))
	if err := cfg.UnmarshalTable(tbl, &x); err != nil {
		t.Fatal(err)
	}
	if x.B != 3 {
		t.Fatalf("wrong value after UnmarshalTable: got %d, want %d", x.B, 3)
	}
}

func TestConfigFieldToKey(t *testing.T) {
	cfg := Config{FieldToKey: func(reflect.Type, string) string { return "A" }}

	x := struct{ B int }{B: 999}
	enc, err := cfg.Marshal(&x)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(enc, []byte("A = 999\n")) {
		t.Fatalf(`got %q, want "A = 999"`, enc)
	}
}

func TestConfigMissingField(t *testing.T) {
	calls := make(map[string]bool)
	cfg := Config{
		NormFieldName: func(rt reflect.Type, field string) string {
			return field
		},
		MissingField: func(rt reflect.Type, field string) error {
			calls[field] = true
			return nil
		},
	}

	var x struct{ B int }
	input := []byte(`
A = 1
B = 2
`)
	if err := cfg.Unmarshal(input, &x); err != nil {
		t.Fatal(err)
	}
	if x.B != 2 {
		t.Errorf("wrong value after Unmarshal: got %d, want %d", x.B, 1)
	}
	if !calls["A"] {
		t.Error("MissingField not called for 'A'")
	}
	if calls["B"] {
		t.Error("MissingField called for 'B'")
	}
}
