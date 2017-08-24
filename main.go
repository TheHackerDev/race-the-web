package main

import (
	"fmt"
	"os"
)

// Main entry function for the program
func main() {
	// Run from command-line if arguments are provided- this means that a configuration file has been provided
	if len(os.Args) >= 2 {
		// Start cmd
		if err, _ := Start(); err != nil {
			fmt.Println(usage)
			outError("[ERROR] %s\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	} else {
		// Start API
		err := StartAPI()
		if err != nil {
			outError("[ERROR] %s\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}
}

// TODO: Write unit tests for all endpoints
