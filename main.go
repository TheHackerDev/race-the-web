package main

import (
	"fmt"
	"log"
	"os"
)

// Main entry function for the program
func main() {
	// Change output location of logs
	log.SetOutput(os.Stdout)

	// Run from command-line if arguments are provided- this means that a configuration file has been provided
	if len(os.Args) >= 2 {
		// Start cmd
		if err := StartCMD(); err != nil {
			fmt.Println(usage)
			outError("[ERROR] %s\n", err)
			os.Exit(1)
		}

	} else {
		// Start API
		StartAPI()
	}
}

// TODO: Write unit tests for all endpoints
