package toml_test

import (
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/naoina/toml"
)

// This example shows how the UnmarshalerRec interface can be used to apply field
// validations and default values.

type Server struct {
	Addr    string        // "<host>:<port>"
	Timeout time.Duration // defaults to 10s
}

// UnmarshalTOML implements toml.Unmarshaler.
func (s *Server) UnmarshalTOML(decode func(interface{}) error) error {
	// Parse the input into type tomlServer, which defines the
	// expected format of Server in TOML.
	type tomlServer struct {
		Addr    string
		Timeout string
	}
	var dec tomlServer
	if err := decode(&dec); err != nil {
		return err
	}

	// Validate the address.
	if dec.Addr == "" {
		return errors.New("missing server address")
	}
	_, _, err := net.SplitHostPort(dec.Addr)
	if err != nil {
		return fmt.Errorf("invalid server address %q: %v", dec.Addr, err)
	}
	// Validate the timeout and apply the default value.
	var timeout time.Duration
	if dec.Timeout == "" {
		timeout = 10 * time.Second
	} else if timeout, err = time.ParseDuration(dec.Timeout); err != nil {
		return fmt.Errorf("invalid server timeout %q: %v", dec.Timeout, err)
	}

	// Assign the decoded value.
	*s = Server{Addr: dec.Addr, Timeout: timeout}
	return nil
}

func ExampleUnmarshalerRec() {
	input := []byte(`
[[servers]]
addr = "198.51.100.3:80"

[[servers]]
addr = "192.0.2.10:8080"
timeout = "30s"
`)
	var config struct {
		Servers []Server
	}
	toml.Unmarshal(input, &config)
	fmt.Printf("Unmarshaled:\n%+v\n\n", config)

	// Output:
	// Unmarshaled:
	// {Servers:[{Addr:198.51.100.3:80 Timeout:10s} {Addr:192.0.2.10:8080 Timeout:30s}]}
}
