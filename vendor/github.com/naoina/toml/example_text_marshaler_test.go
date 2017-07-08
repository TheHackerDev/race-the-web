package toml_test

import (
	"fmt"
	"net"

	"github.com/naoina/toml"
)

func Example_textUnmarshaler() {
	type Config struct {
		Servers []net.IP
	}

	input := []byte(`
servers = ["192.0.2.10", "198.51.100.3"]
`)

	var config Config
	toml.Unmarshal(input, &config)
	fmt.Printf("Unmarshaled:\n%+v\n\n", config)

	output, _ := toml.Marshal(&config)
	fmt.Printf("Marshaled:\n%s", output)

	// Output:
	// Unmarshaled:
	// {Servers:[192.0.2.10 198.51.100.3]}
	//
	// Marshaled:
	// servers = ["192.0.2.10", "198.51.100.3"]
}

func Example_textUnmarshalerError() {
	type Config struct {
		Servers []net.IP
	}

	input := []byte(`
servers = ["192.0.2.10", "198.51.100.500"]
`)

	var config Config
	err := toml.Unmarshal(input, &config)
	fmt.Printf("Unmarshal error:\n%v", err)

	// Output:
	// Unmarshal error:
	// line 2: (toml_test.Config.Servers) invalid IP address: 198.51.100.500
}
