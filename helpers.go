package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
)

// Function SetDefaults sets the default options, if not present in the configuration file.
// Redirects and verbose are both false, as the default value of a boolean.
func SetDefaults(config *Configuration) {
	// Count
	if config.Count == 0 {
		// Set to default value of 100
		config.Count = 100
	}
}

// Function ReadResponseBody is a helper function to read the content from a response's body and refill the body with another io.ReadCloser, so that it can be read again.
func ReadResponseBody(resp *http.Response) (content []byte, err error) {
	// Get the content
	content, err = ioutil.ReadAll(resp.Body)

	// Reset the response body
	rCloser := ioutil.NopCloser(bytes.NewBuffer(content))
	resp.Body = rCloser

	return
}
