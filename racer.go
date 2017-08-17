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
// Defaults:
// Count: 100
// Verbose: false
// Proxy: *none*
type Configuration struct {
	Count   int      `json:"count"`
	Verbose bool     `json:"verbose"`
	Proxy   string   `json:"proxy"`
	Targets []Target `json:"targets" binding:"required"`
}

// Target is a struct to hold information about an individual target URL endpoint.
type Target struct {
	Method    string         `json:"method" binding:"required"`
	URL       string         `json:"url" binding:"required"`
	Body      string         `json:"body"` // NOTE: This is a base64 encoded string of the entire body contents (especially useful in the case of full HTML pages)
	Cookies   []string       `json:"cookies"`
	Headers   []string       `json:"headers"`
	Redirects bool           `json:"redirects"`
	CookieJar http.CookieJar `json:"-"` // Ignore this field, as it is usually nil when outputting via the API
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

// ResponseInfo details information about responses received from targets. Uses *http.Response here for speed, as this struct is used in gathering data quickly back from targets (before comparison begins). This will later be parsed and converted into a UniqueResponseData object.
type ResponseInfo struct {
	Response *http.Response
	Target   Target
}

// UniqueResponseInfo details information about unique responses received from targets
type UniqueResponseInfo struct {
	Response UniqueResponseData
	Targets  []Target
	Count    int
}

// ResponseData is an easily consumable structure holding relevant unique response data
type UniqueResponseData struct {
	Body       string
	StatusCode int
	Length     int64
	Protocol   string
	Headers    http.Header
	Location   string
}

// Usage message
var usage string

// Colour outputs
var outError = color.New(color.FgRed).PrintfFunc()

// Function init initializes the program defaults
func init() {
	usage = fmt.Sprintf("Usage: %s config.toml", os.Args[0])
}

// Function Start is the entrypoint to the race testing component of the application. It sends the work to the appropriate functions, sequentially.
// Start also handles logging for the race tests.
func Start() (error, []UniqueResponseInfo) {
	// Change output location of logs
	log.SetOutput(os.Stdout)

	if len(os.Args) >= 2 {
		// Check the config file
		configFile := os.Args[1]
		var err error
		configuration, err = getConfigFile(configFile)
		if err != nil {
			return err, nil
		}
	}

	// Verify that config is present
	if len(configuration.Targets) == 0 {
		// No targets specified
		return fmt.Errorf("No targets set. Minimum of 1 target required."), nil
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

	// Output the responses
	outputResponses(uniqueResponses)

	// Return the responses back to the API
	return nil, uniqueResponses
}

// Function getConfigFile checks that all necessary configuration fields are given
// in a valid config file, and parses it for data.
// Returns a Configuration object if successful.
// Returns an empty Configuration object and a custom error if something went wrong.
func getConfigFile(location string) (Configuration, error) {
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

	// Set default values
	SetDefaults(&config)

	return config, nil
}

// Prepares an attack by parsing a global configuration.
// Returns an error if something went wrong.
func prepareAttack() error {
	// Add the cookies to the cookiejar for each target
	for _, target := range configuration.Targets {
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
			return fmt.Errorf("Error parsing target URL: %s", err.Error())
		}
		target.CookieJar.SetCookies(targetURL, cookies)
	}

	// Set a proxy for all http requests, if specified
	if configuration.Proxy != "" {
		proxyURL, err := url.Parse(configuration.Proxy)
		if err != nil {
			return fmt.Errorf("Invalid proxy URL.")
		}
		if proxyURL.Scheme == "" {
			proxyURL.Scheme = "http" // default of http
			configuration.Proxy = proxyURL.String()
		} else if proxyURL.Scheme != "http" && proxyURL.Scheme != "https" {
			return fmt.Errorf("Proxy must be an http or https proxy, and specify the proper scheme (e.g. \"http://127.0.0.1:8080\")")
		}
	}

	return nil
}

// Function SetDefaults sets the default options, if not present in the configuration file.
// Redirects and verbose are both false, as the default value of a boolean.
func SetDefaults(config *Configuration) {
	// Count
	if config.Count == 0 {
		// Set to default value of 100
		config.Count = 100
	}
}

// Function sendRequests takes care of sending the requests to the target concurrently.
// Errors are passed back in a channel of errors. If the length is zero, there were no errors.
func sendRequests() (responses chan ResponseInfo, errors chan error) {
	// Initialize the concurrency objects
	responses = make(chan ResponseInfo, configuration.Count*len(configuration.Targets))
	errors = make(chan error, configuration.Count*len(configuration.Targets))
	urlsInProgress.Add(configuration.Count * len(configuration.Targets))

	// Send requests to multiple URLs (if present) the same number of times
	for _, target := range configuration.Targets {
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
				if configuration.Proxy != "" {
					log.Printf("[VERBOSE] Proxy: %s\n", configuration.Proxy)
				}
				if t.Body != "" {
					log.Printf("[VERBOSE] Request body: %s\n", t.Body)
				}
				if len(t.Cookies) > 0 {
					log.Printf("[VERBOSE] Request cookies: %v\n", t.Cookies)
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

					// Add custom headers to the request
					for _, header := range t.Headers {
						split := strings.Split(header, ":")
						hKey := split[0]
						hVal := split[1]
						req.Header.Add(hKey, hVal)
					}

					// Add content-type to POST requests (some applications require this to properly process POST requests)
					// TODO: Find any bugs around other request types
					if t.Method == "POST" {
						req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

					}

					// TODO: Add context to http client requests to manually specify timeout options (new in Go 1.7)

					var transport http.Transport
					// Use proxy, if set
					if configuration.Proxy != "" {
						proxyURL, _ := url.Parse(configuration.Proxy) // error checked when getting configuration
						transport = http.Transport{
							TLSClientConfig: &tls.Config{
								InsecureSkipVerify: true,
							},
							Proxy: http.ProxyURL(proxyURL),
						}
					} else {
						transport = http.Transport{
							TLSClientConfig: &tls.Config{
								InsecureSkipVerify: true,
							},
						}
					}

					if t.Redirects {
						client = http.Client{
							Jar:       t.CookieJar,
							Transport: &transport,
							Timeout:   120 * time.Second,
						}
					} else {
						client = http.Client{
							Jar:       t.CookieJar,
							Transport: &transport,
							CheckRedirect: func(req *http.Request, via []*http.Request) error {
								// Craft the custom error
								redirectError := RedirectError{req}
								return &redirectError
							},
							Timeout: 120 * time.Second,
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

	// Close the response and error channels, so they don't block on the range read
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

		// Create response data object to pass around
		respData := UniqueResponseData{
			Body:       string(respBody),
			StatusCode: respInfo.Response.StatusCode,
			Length:     respInfo.Response.ContentLength,
			Protocol:   respInfo.Response.Proto,
			Headers:    respInfo.Response.Header}
		location, err := respInfo.Response.Location()
		if err != http.ErrNoLocation {
			respData.Location = location.String()
		}

		if len(uniqueResponses) == 0 {
			// The unique responses slice is empty, add the current response as the first
			uniqueResponses = append(uniqueResponses, UniqueResponseInfo{
				Count:    1,
				Response: respData,
				Targets:  []Target{respInfo.Target}})
			continue
		}

		// Add to the unique responses channel, if no similar ones exist
		respMatch := false        // Assume unique, until similar found
		j := len(uniqueResponses) // Used to count through the existing unique responses channel
		for i := 0; i < j; i++ {
			compareResp := &uniqueResponses[i]

			// Compare response status code, body content, and content length
			if respData.StatusCode == compareResp.Response.StatusCode && respData.Body == compareResp.Response.Body && respData.Length == compareResp.Response.Length {
				// Match found
				respMatch = true
				compareResp.Count++

				// Check for the same request that generated this matched response (== unique request AND response)
				reqMatch := false
				// Iterate through all requests in comparison group and compare against current request being processed
				for _, compareTarget := range compareResp.Targets {
					if reflect.DeepEqual(compareTarget, respInfo.Target) {
						// Request match found
						reqMatch = true
						break
					}
				}
				if !reqMatch {
					// Append the new target to the unique response
					compareResp.Targets = append(compareResp.Targets, respInfo.Target)
				}
				// Exit inner loop
				break
			}
		}

		// Check if response matches another response already found
		if !respMatch {
			// Unique, add to unique responses
			uniqueResponses = append(uniqueResponses, UniqueResponseInfo{
				Count:    1,
				Response: respData,
				Targets:  []Target{respInfo.Target}})
			// Increase loop count to account for newly added unique response
			j++
		}
	}

	// VERBOSE
	if configuration.Verbose {
		log.Printf("[VERBOSE] Unique response comparison complete.\n")
	}

	// Close the channels
	close(errors)

	return
}

func outputResponses(uniqueResponses []UniqueResponseInfo) {
	// Display the responses
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
