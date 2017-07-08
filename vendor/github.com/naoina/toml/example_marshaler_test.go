package toml_test

import (
	"fmt"

	"github.com/naoina/toml"
)

// This example shows how the Unmarshaler interface can be used to access the TOML source
// input during decoding.

// RawTOML stores the raw TOML syntax that was passed to UnmarshalTOML.
type RawTOML []byte

func (r *RawTOML) UnmarshalTOML(input []byte) error {
	cpy := make([]byte, len(input))
	copy(cpy, input)
	*r = cpy
	return nil
}

func ExampleUnmarshaler() {
	input := []byte(`
foo = 1

[[servers]]
addr = "198.51.100.3:80" # a comment

[[servers]]
addr = "192.0.2.10:8080"
timeout = "30s"
`)
	var config struct {
		Foo     int
		Servers RawTOML
	}
	toml.Unmarshal(input, &config)
	fmt.Printf("config.Foo = %d\n", config.Foo)
	fmt.Printf("config.Servers =\n%s\n", indent(config.Servers, 2))

	// Output:
	// config.Foo = 1
	// config.Servers =
	//   [[servers]]
	//   addr = "198.51.100.3:80" # a comment
	//   [[servers]]
	//   addr = "192.0.2.10:8080"
	//   timeout = "30s"
}
