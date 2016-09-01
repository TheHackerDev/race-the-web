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
	"strings"
	"sync"
	"time"

	// Used to parse TOML configuration file
	"github.com/naoina/toml"
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

type Target struct {
	Method    string
	Url       string
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

// Usage message
var usage string

// Function init initializes the program defaults
func init() {
	// TODO: set the usage string
	usage = ``
}

// Function main is the entrypoint to the application. It sends the work to the appropriate functions, sequentially.
func main() {
	// Change output location of logs
	log.SetOutput(os.Stdout)

	// Check the config file
	if len(os.Args) != 2 {
		// No configuration file provided
		log.Println("[ERROR] No configuration file location provided.")
		fmt.Println(usage)
		os.Exit(1)
	}
	configFile := os.Args[1]
	var err error
	configuration, err = getConfig(configFile)
	if err != nil {
		log.Println(err.Error())
		fmt.Println(usage)
		os.Exit(1)
	}

	// Send the requests concurrently
	log.Println("Requests begin.")
	responses, errors := sendRequests()
	if len(errors) != 0 {
		for err := range errors {
			log.Printf("[ERROR] %s\n", err.Error())
		}
	}

	// Make sure all response bodies are closed- memory leaks otherwise
	defer func() {
		for resp := range responses {
			resp.Body.Close()
		}
	}()

	// Compare the responses for uniqueness
	uniqueResponses, errors := compareResponses(responses)
	if len(errors) != 0 {
		for err := range errors {
			log.Printf("[ERROR] %s\n", err.Error())
		}
	}

	// Make sure all response bodies are closed- memory leaks otherwise
	defer func() {
		for resp := range uniqueResponses {
			resp.Body.Close()
		}
	}()

	// Output the responses
	outputResponses(uniqueResponses)

	// Echo completion
	log.Println("Complete.")
}

// Function getConfig checks that all necessary configuration fields are given
// in a valid config file, and parses it for data.
// Returns a Configuration object if successful.
// Returns an empty Configuration object and a custom error if something went wrong.
func getConfig(location string) (Configuration, error) {
	f, err := os.Open(location)
	if err != nil {
		return Configuration{}, fmt.Errorf("[ERROR] Error opening configuration file: %s", err.Error())
	}
	defer f.Close()

	buf, err := ioutil.ReadAll(f)
	if err != nil {
		return Configuration{}, fmt.Errorf("[ERROR] Error reading from configuration file: %s", err.Error())
	}
	var config Configuration
	// Parse all data from the provided configuration file into a Configuration object
	if err := toml.Unmarshal(buf, &config); err != nil {
		return Configuration{}, fmt.Errorf("[ERROR] Error with TOML file: %s", err.Error())
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
		targetUrl, err := url.Parse(target.Url)
		if err != nil {
			return Configuration{}, fmt.Errorf("[ERROR] Error parsing target URL: %s", err.Error())
		}
		target.CookieJar.SetCookies(targetUrl, cookies)
	}

	// TODO: Add defaults here, if values not present
	// Redirects = false
	// Count = 100
	// Verbose = false

	// TODO: Check if targets empty, if so, return with an error

	return config, nil
}

// Function sendRequests takes care of sending the requests to the target concurrently.
// Errors are passed back in a channel of errors. If the length is zero, there were no errors.
func sendRequests() (responses chan *http.Response, errors chan error) {
	// Initialize the concurrency objects
	responses = make(chan *http.Response, configuration.Count*len(configuration.Target))
	errors = make(chan error, configuration.Count*len(configuration.Target))
	urlsInProgress.Add(configuration.Count * len(configuration.Target))

	// Send requests to multiple URLs (if present) the same number of times
	for _, target := range configuration.Target {
		go func(t Target) {
			// Cast the target URL to a URL type
			tUrl, err := url.Parse(t.Url)
			if err != nil {
				errors <- fmt.Errorf("[ERROR] Error parsing URL %s: %v", t.Url, err.Error())
				return
			}

			// VERBOSE
			if configuration.Verbose {
				log.Printf("[VERBOSE] Sending %d %s requests to %s\n", configuration.Count, t.Method, tUrl.String())
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
					req, err := http.NewRequest(t.Method, tUrl.String(), requestBody)
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
								responses <- resp
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
						responses <- resp
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
func compareResponses(responses chan *http.Response) (uniqueResponses map[*http.Response]int, errors chan error) {
	// Initialize the unique responses map
	uniqueResponses = make(map[*http.Response]int)

	// Initialize the error channel
	errors = make(chan error, len(responses))

	// VERBOSE
	if configuration.Verbose {
		log.Printf("[VERBOSE] Unique response comparison begin.\n")
	}

	// Compare the responses, one at a time
	for resp := range responses {
		// Read the response body
		respBody, err := readResponseBody(resp)
		if err != nil {
			errors <- fmt.Errorf("Error reading response body: %s", err.Error())

			// Exit this loop
			continue
		}

		// Add an entry, if the unique responses map is empty
		if len(uniqueResponses) == 0 {
			uniqueResponses[resp] = 0
		} else {
			// Add to the unique responses map, if no similar ones exist
			// Assume unique, until similar found
			unique := true
			for uResp := range uniqueResponses {
				// Read the unique response body
				uRespBody, err := readResponseBody(uResp)
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
				if resp.StatusCode == uResp.StatusCode && resp.ContentLength == uResp.ContentLength && respBodyMatch {
					// Match, increase count
					uniqueResponses[uResp]++
					unique = false
					// Exit inner loop
					break
				}
			}

			// Check if unique from all other unique responses
			if unique {
				// Unique, add to unique responses
				uniqueResponses[resp] = 0
			}
		}
	}

	// VERBOSE
	if configuration.Verbose {
		log.Printf("[VERBOSE] Unique response comparision complete.\n")
	}

	// Close the error channel
	close(errors)

	return
}

func outputResponses(uniqueResponses map[*http.Response]int) {
	// Display the responses
	log.Printf("Responses:\n")
	for resp, count := range uniqueResponses {
		// TODO: Output request here
		fmt.Printf("Response:\n")
		fmt.Printf("[Status Code] %v\n", resp.StatusCode)
		fmt.Printf("[Protocol] %v\n", resp.Proto)
		if len(resp.Header) != 0 {
			fmt.Println("[Headers]")
			for header, value := range resp.Header {
				fmt.Printf("\t%v: %v\n", header, value)
			}
		}
		location, err := resp.Location()
		if err != http.ErrNoLocation {
			fmt.Printf("[Location] %v\n", location.String())
		}
		respBody, err := readResponseBody(resp)
		if err != nil {
			fmt.Println("[Body] ")
			fmt.Printf("[ERROR] Error reading body: %v.", err)
		} else {
			fmt.Printf("[Body]\n%s\n", respBody)
			// Close the response body
			resp.Body.Close()
		}
		fmt.Printf("Similar: %v\n\n", count)
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
