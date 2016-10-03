package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"

	// Used to parse TOML configuration file
	"github.com/naoina/toml"
	// Used to output in colour to the console
	"github.com/fatih/color"
)

// urlsInProgress is a wait group, for concurrency
var urlsInProgress sync.WaitGroup

// RedirectError is a custom error type for following redirects, and can be safely ignored
type RedirectError struct {
	RedirectRequest *http.Request
}

// Error method returns a string of the error
func (err *RedirectError) Error() string {
	return fmt.Sprintf("Redirect not followed to: %v", err.RedirectRequest.URL.String())
}

// Configuration holds all the configuration data passed in from the config.TOML file.
type Configuration struct {
	Count   int
	Verbose bool
	Target  []Target
}

// Target is a struct to hold information about an individual target URL endpoint.
type Target struct {
	Method    string
	URL       string
	Body      string
	Cookies   []string
	Redirects bool
	CookieJar http.CookieJar
}

// REF: Access parts of the Configuration object.
// fmt.Printf("All: %v\n", config)
// fmt.Printf("Count: %v\n", config.Count)
// fmt.Printf("Verbose: %v\n", config.Verbose)
// fmt.Println("Targets:")
// for _, target := range config.Target {
// 	fmt.Printf("\n\tMethod: %s\n", target.Method)
// 	fmt.Printf("\tURL: %s\n", target.Url)
// 	fmt.Printf("\tBody: %s\n", target.Body)
// 	fmt.Printf("\tRedirects: %v\n", target.Redirects)
// 	for _, cookie := range target.Cookies {
// 		fmt.Printf("\tCookie: %s\n", cookie)
// 	}
// 	// Add the cookie jar after TOML is unmarshaled
// 	target.CookieJar, _ = cookiejar.New(nil)
// 	fmt.Printf("\tCookieJar: %v\n", target.CookieJar)
// }

var configuration Configuration

// ResponseInfo details information about responses received from targets
type ResponseInfo struct {
	Response *http.Response
	Target   Target
}

// UniqueResponseInfo details information about unique responses received from targets
type UniqueResponseInfo struct {
	Response *http.Response
	Targets  []Target
	Count    int
}

// Usage message
var usage string

// Colour outputs
var outError = color.New(color.FgRed).PrintfFunc()

// Function init initializes the program defaults
func init() {
	usage = fmt.Sprintf("Usage: %s config.toml", os.Args[0])
}

// Function main is the entrypoint to the application. It sends the work to the appropriate functions, sequentially.
func main() {
	// Change output location of logs
	log.SetOutput(os.Stdout)

	// Check the config file
	if len(os.Args) != 2 {
		// No configuration file provided
		outError("[ERROR] No configuration file location provided.")
		fmt.Println(usage)
		os.Exit(1)
	}
	configFile := os.Args[1]
	var err error
	configuration, err = getConfig(configFile)
	if err != nil {
		outError("[ERROR] %s\n", err.Error())
		fmt.Println(usage)
		os.Exit(1)
	}

	// Send the requests concurrently
	log.Println("Requests begin.")
	responses, errors := sendRequests()
	if len(errors) != 0 {
		for err := range errors {
			outError("[ERROR] %s\n", err.Error())
		}
	}
	log.Println("Requests completed.")

	// Make sure all response bodies are closed- memory leaks otherwise
	defer func() {
		for respInfo := range responses {
			respInfo.Response.Body.Close()
		}
	}()

	// Compare the responses for uniqueness
	uniqueResponses, errors := compareResponses(responses)
	if len(errors) != 0 {
		for err := range errors {
			outError("[ERROR] %s\n", err.Error())
		}
	}

	// Make sure all response bodies are closed- memory leaks otherwise
	defer func() {
		for _, uRespInfo := range uniqueResponses {
			uRespInfo.Response.Body.Close()
		}
	}()

	// Output the responses
	outputResponses(uniqueResponses)
}

// Function getConfig checks that all necessary configuration fields are given
// in a valid config file, and parses it for data.
// Returns a Configuration object if successful.
// Returns an empty Configuration object and a custom error if something went wrong.
func getConfig(location string) (Configuration, error) {
	f, err := os.Open(location)
	if err != nil {
		return Configuration{}, fmt.Errorf("Error opening configuration file: %s", err.Error())
	}
	defer f.Close()

	buf, err := ioutil.ReadAll(f)
	if err != nil {
		return Configuration{}, fmt.Errorf("Error reading from configuration file: %s", err.Error())
	}
	var config Configuration
	// Parse all data from the provided configuration file into a Configuration object
	if err := toml.Unmarshal(buf, &config); err != nil {
		return Configuration{}, fmt.Errorf("Error with TOML file: %s", err.Error())
	}

	// Add the cookies to the cookiejar for each target
	for _, target := range config.Target {
		target.CookieJar, _ = cookiejar.New(nil)
		var cookies []*http.Cookie
		for _, c := range target.Cookies {
			// Split the cookie name and value
			vals := strings.Split(c, "=")
			cookieName := strings.TrimSpace(vals[0])
			cookieValue := strings.TrimSpace(vals[1])

			// Create the cookie
			cookie := &http.Cookie{
				Name:  cookieName,
				Value: cookieValue,
			}

			// Add the cookie to the new slice of cookies
			cookies = append(cookies, cookie)
		}

		// Associate the cookies with the current target
		targetURL, err := url.Parse(target.URL)
		if err != nil {
			return Configuration{}, fmt.Errorf("Error parsing target URL: %s", err.Error())
		}
		target.CookieJar.SetCookies(targetURL, cookies)
	}

	// Set default values
	config = setDefaults(config)

	if len(config.Target) == 0 {
		// No targets specified
		return Configuration{}, fmt.Errorf("No targets set. Minimum of 1 target required.")
	}

	return config, nil
}

// Function setDefaults sets the default options, if not present in the configuration file.
// Redirects and verbose are both false, as the default value of a boolean.
func setDefaults(config Configuration) Configuration {
	// Count
	if config.Count == 0 {
		// Set to default value of 100
		config.Count = 100
	}

	return config
}

// Function sendRequests takes care of sending the requests to the target concurrently.
// Errors are passed back in a channel of errors. If the length is zero, there were no errors.
func sendRequests() (responses chan ResponseInfo, errors chan error) {
	// Initialize the concurrency objects
	responses = make(chan ResponseInfo, configuration.Count*len(configuration.Target))
	errors = make(chan error, configuration.Count*len(configuration.Target))
	urlsInProgress.Add(configuration.Count * len(configuration.Target))

	// Send requests to multiple URLs (if present) the same number of times
	for _, target := range configuration.Target {
		go func(t Target) {
			// Cast the target URL to a URL type
			tURL, err := url.Parse(t.URL)
			if err != nil {
				errors <- fmt.Errorf("Error parsing URL %s: %v", t.URL, err.Error())
				return
			}

			// VERBOSE
			if configuration.Verbose {
				log.Printf("[VERBOSE] Sending %d %s requests to %s\n", configuration.Count, t.Method, tURL.String())
				if t.Body != "" {
					log.Printf("[VERBOSE] Request body: %s", t.Body)
				}
			}
			for i := 0; i < configuration.Count; i++ {
				go func(index int) {
					// Ensure that the waitgroup element is returned
					defer urlsInProgress.Done()

					// Convert the request body to an io.Reader interface, to pass to the request.
					// This must be done in the loop, because any call to client.Do() will
					// read the body contents on the first time, but not any subsequent requests.
					requestBody := strings.NewReader(t.Body)

					// Declare HTTP request method and URL
					req, err := http.NewRequest(t.Method, tURL.String(), requestBody)
					if err != nil {
						errors <- fmt.Errorf("Error in forming request: %v", err.Error())
						return
					}

					// Create the HTTP client
					// Using Cookie jar
					// Ignoring TLS errors
					// Ignoring redirects (more accurate output), depending on user flag
					// Implementing a connection timeouts, for slow clients & servers (especially important with race conditions on the server)
					var client http.Client

					// TEMP- append cookies directly to the request
					if len(t.Cookies) > 0 {
						cookieStr := strings.Join(t.Cookies, ";")
						req.Header.Add("Cookie", cookieStr)
					}

					// Add content-type to POST requests (some applications require this to properly process POST requests)
					// TODO: Find any bugs around other request types
					if t.Method == "POST" {
						req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

					}

					// TODO: Add context to http client requests to manually specify timeout options (new in Go 1.7)

					if t.Redirects {
						client = http.Client{
							Jar: t.CookieJar,
							Transport: &http.Transport{
								TLSClientConfig: &tls.Config{
									InsecureSkipVerify: true,
								},
							},
							Timeout: 20 * time.Second,
						}
					} else {
						client = http.Client{
							Jar: t.CookieJar,
							Transport: &http.Transport{
								TLSClientConfig: &tls.Config{
									InsecureSkipVerify: true,
								},
							},
							CheckRedirect: func(req *http.Request, via []*http.Request) error {
								// Craft the custom error
								redirectError := RedirectError{req}
								return &redirectError
							},
							Timeout: 20 * time.Second,
						}
					}

					// Make the request
					resp, err := client.Do(req)
					// Check the error type from the request
					if err != nil {
						if uErr, ok := err.(*url.Error); ok {
							if rErr, ok2 := uErr.Err.(*RedirectError); ok2 {
								// Redirect error
								// VERBOSE
								if configuration.Verbose {
									log.Printf("[VERBOSE] %v\n", rErr)
								}
								// Add the response to the responses channel, because it is still valid
								responses <- ResponseInfo{Response: resp, Target: t}
							} else {
								// URL Error, but not a redirect error
								errors <- fmt.Errorf("Error in request #%v: %v\n", index, err)
							}
						} else {
							// Other type of error
							errors <- fmt.Errorf("Error in request #%v: %v\n", index, err)
						}
					} else {
						// Add the response to the responses channel
						responses <- ResponseInfo{Response: resp, Target: t}
					}
				}(i)
			}
		}(target)
	}

	// Wait for the URLs to finish sending
	urlsInProgress.Wait()

	// VERBOSE
	if configuration.Verbose {
		log.Printf("[VERBOSE] Requests complete.")
	}

	// Close the response and error chanels, so they don't block on the range read
	close(responses)
	close(errors)

	return
}

// Function compareResponses compares the responses returned from the requests,
// and adds them to a map, where the key is an *http.Response, and the value is
// the number of similar responses observed.
func compareResponses(responses chan ResponseInfo) (uniqueResponses []UniqueResponseInfo, errors chan error) {
	// Initialize the channels
	errors = make(chan error, len(responses))

	// VERBOSE
	if configuration.Verbose {
		log.Printf("[VERBOSE] Unique response comparison begin.\n")
	}

	// Compare the responses, one at a time
	for respInfo := range responses {
		// Read the response body
		respBody, err := readResponseBody(respInfo.Response)
		if err != nil {
			errors <- fmt.Errorf("Error reading response body: %s", err.Error())

			// Exit this loop
			continue
		}

		if len(uniqueResponses) == 0 {
			// The unique responses slice is empty, add the current response as the first
			uniqueResponses = append(uniqueResponses, UniqueResponseInfo{Count: 1, Response: respInfo.Response, Targets: []Target{respInfo.Target}})
		} else {
			// Add to the unique responses channel, if no similar ones exist
			match := false // Assume unique, until similar found
			j := len(uniqueResponses)
			for i := 0; i < j; i++ {
				uRespInfo := &uniqueResponses[i]
				// Read the unique response body
				uRespBody, err := readResponseBody(uRespInfo.Response)
				if err != nil {
					errors <- fmt.Errorf("Error reading unique response body: %s", err.Error())

					// Error, move on to the next inner loop value
					continue
				}

				// Compare the response bodies
				respBodyMatch := false
				if string(respBody) == string(uRespBody) {
					respBodyMatch = true
				}

				// Compare response status code, body content, and content length
				if respInfo.Response.StatusCode == uRespInfo.Response.StatusCode && respInfo.Response.ContentLength == uRespInfo.Response.ContentLength && respBodyMatch {
					// Match found
					match = true
					uRespInfo.Count++

					// Check for the same request, using the target information
					targetMatch := false
					for _, target := range uRespInfo.Targets {
						if reflect.DeepEqual(target, respInfo.Target) {
							// Target match found
							targetMatch = true
							break
						}
					}
					if !targetMatch {
						// Append the new target to the unique response
						uRespInfo.Targets = append(uRespInfo.Targets, respInfo.Target)
					}
					// Exit inner loop
					break
				}
			}

			// Check if response matches another response already found
			if !match {
				// Unique, add to unique responses
				uniqueResponses = append(uniqueResponses, UniqueResponseInfo{Count: 1, Response: respInfo.Response, Targets: []Target{respInfo.Target}})
				// Increase loop count to account for newly added unique response
				j++
			}
		}
	}

	// VERBOSE
	if configuration.Verbose {
		log.Printf("[VERBOSE] Unique response comparision complete.\n")
	}

	// Close the channels
	close(errors)

	return
}

func outputResponses(uniqueResponses []UniqueResponseInfo) {
	// Display the responses
	fmt.Printf("Unique Responses:\n\n")
	for _, uRespInfo := range uniqueResponses {
		fmt.Println("**************************************************")
		fmt.Printf("RESPONSE:\n")
		fmt.Printf("[Status Code] %v\n", uRespInfo.Response.StatusCode)
		fmt.Printf("[Protocol] %v\n", uRespInfo.Response.Proto)
		if len(uRespInfo.Response.Header) != 0 {
			fmt.Println("[Headers]")
			for header, value := range uRespInfo.Response.Header {
				fmt.Printf("\t%v: %v\n", header, value)
			}
		}
		location, err := uRespInfo.Response.Location()
		if err != http.ErrNoLocation {
			fmt.Printf("[Location] %v\n", location.String())
		}
		respBody, err := readResponseBody(uRespInfo.Response)
		if err != nil {
			fmt.Println("[Body] ")
			outError("[ERROR] Error reading body: %v.", err)
		} else {
			fmt.Printf("[Body]\n%s\n", respBody)
			// Close the response body
			uRespInfo.Response.Body.Close()
		}
		fmt.Printf("Similar: %v\n", uRespInfo.Count-1)
		fmt.Printf("REQUESTS:\n")
		for _, target := range uRespInfo.Targets {
			fmt.Printf("\tURL: %s\n", target.URL)
			fmt.Printf("\tMethod: %s\n", target.Method)
			fmt.Printf("\tBody: %s\n", target.Body)
			fmt.Printf("\tCookies: %v\n", target.Cookies)
			fmt.Printf("\tRedirects: %t\n", target.Redirects)
			fmt.Println()
		}
	}
}

// Function readResponseBody is a helper function to read the content form a response's body,
// and refill the body with another io.ReadCloser, so that it can be read again.
func readResponseBody(resp *http.Response) (content []byte, err error) {
	// Get the content
	content, err = ioutil.ReadAll(resp.Body)

	// Reset the response body
	rCloser := ioutil.NopCloser(bytes.NewBuffer(content))
	resp.Body = rCloser

	return
}

// TODO: Add option for output (boolean), which toggles whether to compare responses and display responses in the output.
// Useful in cases where there is another way to validate the race condition (in the bank test application, for example)
