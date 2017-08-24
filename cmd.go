package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/naoina/toml"
)

// StartCMD begins the program with command-line usage.
// Returns any errors encountered during operation.
func StartCMD() error {
	// Check the config file
	configFile := os.Args[1]
	var err error
	configuration, err = getConfigFile(configFile)
	if err != nil {
		return err
	}

	// Set default values
	SetDefaults(&configuration)

	// Run the race test
	err, responseData := StartRace()
	if err != nil {
		return err
	}

	// Output responses
	outputResponses(responseData)

	return nil
}

// Function getConfigFile checks that all necessary configuration fields are given
// in a valid config file, and parses it for data.
// Returns a Configuration object if successful.
// Returns an empty Configuration object and a custom error if something went wrong.
func getConfigFile(location string) (Configuration, error) {
	f, err := os.Open(location)
	if err != nil {
		return Configuration{}, fmt.Errorf("could not open configuration file: %s", err.Error())
	}
	defer f.Close()

	buf, err := ioutil.ReadAll(f)
	if err != nil {
		return Configuration{}, fmt.Errorf("could not read configuration file: %s", err.Error())
	}
	var config Configuration
	// Parse all data from the provided configuration file into a Configuration object
	if err := toml.Unmarshal(buf, &config); err != nil {
		return Configuration{}, fmt.Errorf("could not unmarshal TOML file: %s", err.Error())
	}

	return config, nil
}

// outputResponses logs the response data to the command line
func outputResponses(uniqueResponses []UniqueResponseInfo) {
	fmt.Printf("Unique Responses:\n\n")
	for _, data := range uniqueResponses {
		fmt.Println("**************************************************")
		fmt.Printf("RESPONSE:\n")
		fmt.Printf("[Status Code] %v\n", data.Response.StatusCode)
		fmt.Printf("[Protocol] %v\n", data.Response.Protocol)
		if len(data.Response.Headers) != 0 {
			fmt.Println("[Headers]")
			for header, value := range data.Response.Headers {
				fmt.Printf("\t%v: %v\n", header, value)
			}
		}
		fmt.Printf("[Location] %v\n", data.Response.Location)
		fmt.Printf("[Body]\n%s\n", data.Response.Body)
		fmt.Printf("Similar: %v\n", data.Count-1)
		fmt.Printf("REQUESTS:\n")
		for _, target := range data.Targets {
			fmt.Printf("\tURL: %s\n", target.URL)
			fmt.Printf("\tMethod: %s\n", target.Method)
			fmt.Printf("\tBody: %s\n", target.Body)
			fmt.Printf("\tCookies: %v\n", target.Cookies)
			if configuration.Proxy != "" {
				fmt.Printf("\tProxy: %v\n", configuration.Proxy)
			}
			fmt.Printf("\tRedirects: %t\n", target.Redirects)
			fmt.Println()
		}
	}
}
