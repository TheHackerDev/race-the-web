package toml_test

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/naoina/toml"
)

type Config struct {
	Title string
	Owner struct {
		Name string
		Org  string `toml:"organization"`
		Bio  string
		Dob  time.Time
	}
	Database struct {
		Server        string
		Ports         []int
		ConnectionMax uint
		Enabled       bool
	}
	Servers map[string]ServerInfo
	Clients struct {
		Data  [][]interface{}
		Hosts []string
	}
}

type ServerInfo struct {
	IP net.IP
	DC string
}

func Example() {
	f, err := os.Open("testdata/example.toml")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	var config Config
	if err := toml.NewDecoder(f).Decode(&config); err != nil {
		panic(err)
	}

	fmt.Println("IP of server 'alpha':", config.Servers["alpha"].IP)
	// Output: IP of server 'alpha': 10.0.0.1
}
