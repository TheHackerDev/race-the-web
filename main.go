package main

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"flag"
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

// Request body content
var body string

// Target URL value
var targetURL *url.URL

// Cookie jar value
var jar *cookiejar.Jar

// Request type
var requestMethod string

// Follow redirects
var followRedirects bool

// Command-line flags
var flagTargetURL = flag.String("url", "", "URL to send the request to.")
var flagBodyFile = flag.String("body", "", "The location (relative or absolute path) of a file containing the body of the request.")
var flagCookiesFile = flag.String("cookies", "", "The location (relative or absolute path) of a file containing newline-separate cookie values being sent along with the request. Cookie names and values are separated by a comma. For example: cookiename,cookieval")
var flagNumRequests = flag.Int("requests", 100, "The number of requests to send to the destination URL.")
var flagRequestMethod = flag.String("method", "POST", "The request type. Can be either `POST, GET, HEAD, PUT`.")
var flagFollowRedirects = flag.Bool("redirects", false, "Follow redirects (3xx status code in responses)")
var flagVerbose = flag.Bool("v", false, "Enable verbose logging.")

func main() {
	// Change output location of logs
	log.SetOutput(os.Stdout)

	// Check the flags
	err := checkFlags()
	if err != nil {
		log.Println(err.Error())
		flag.Usage()
		os.Exit(1)
	}

	// Send the requests concurrently
	log.Println("Requests begin.")
	responses, errors := sendRequests(*flagNumRequests)
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

// Function checkFlags checks that all necessary flags are entered, and parses them for contents.
// Returns a custom error if something went wrong.
func checkFlags() error {
	// Parse the flags
	flag.Parse()

	// Determine whether to follow redirects
	followRedirects = *flagFollowRedirects

	// Set the request type
	switch strings.ToUpper(*flagRequestMethod) {
	case "POST":
		requestMethod = "POST"
	case "GET":
		requestMethod = "GET"
	case "PUT":
		requestMethod = "PUT"
	case "HEAD":
		requestMethod = "HEAD"
	default:
		// Invalid request type specified
		return fmt.Errorf("Invalid request type specified.")
	}

	// Ensure that the destination URL is present
	if *flagTargetURL == "" {
		return fmt.Errorf("Destination URL required.")
	}

	// Parse the URL
	var err error
	targetURL, err = url.Parse(*flagTargetURL)
	if err != nil {
		return fmt.Errorf("Invalid URL provided: %s", *flagTargetURL)
	}

	// Get the request body content
	if *flagBodyFile != "" {
		buf, err := ioutil.ReadFile(*flagBodyFile)
		if err != nil {
			// Error opening the file
			return fmt.Errorf("Unable to open the file: %s\n", *flagBodyFile)
		}
		body = string(buf)
	} else {
		// Body file flag not present, exit.
		return fmt.Errorf("Request body contents required.")
	}

	// Initialize the cookie jar
	jar, _ = cookiejar.New(nil)
	var cookies []*http.Cookie
	// Get the cookies to pass to the request
	if *flagCookiesFile != "" {
		file, err := os.Open(*flagCookiesFile)
		if err != nil {
			// Error opening the file
			return fmt.Errorf("Unable to open the file: %s", *flagCookiesFile)
		}

		// Ensure the file is closed
		defer file.Close()

		// Initialize the file scanner
		scanner := bufio.NewScanner(file)

		// Iterate through the file to get the cookies
		for scanner.Scan() {
			// Parse the line to separate the cookie names and values
			nextLine := scanner.Text()
			vals := strings.Split(nextLine, ",")
			cookieName := strings.TrimSpace(vals[0])
			cookieValue := strings.TrimSpace(vals[1])

			// Create the cookie
			cookie := &http.Cookie{
				Name:  cookieName,
				Value: cookieValue,
			}

			// Add the cookie to the existing slice of cookies
			cookies = append(cookies, cookie)
		}

		// Set the cookies to the appropriate URL
		jar.SetCookies(targetURL, cookies)

	}

	// Made it through with no errors, return
	return nil
}

// Function sendRequests takes care of sending the requests to the target concurrently.
// Errors are passed back in a channel of errors. If the length is zero, there were no errors.
func sendRequests(numRequests int) (responses chan *http.Response, errors chan error) {
	// Initialize the concurrency objects
	responses = make(chan *http.Response, numRequests)
	errors = make(chan error, numRequests)
	urlsInProgress.Add(numRequests)

	// VERBOSE
	if *flagVerbose {
		log.Printf("[VERBOSE] Sending %d %s requests to %s\n", numRequests, requestMethod, targetURL.String())
		if body != "" {
			log.Printf("[VERBOSE] Request body: %s", body)
		}
	}
	for i := 0; i < numRequests; i++ {
		go func(index int) {
			// Ensure that the waitgroup element is returned
			defer urlsInProgress.Done()

			// Convert the request body to an io.Reader interface, to pass to the request.
			// This must be done in the loop, because any call to client.Do() will
			// read the body contents on the first time, but not any subsequent requests.
			requestBody := strings.NewReader(body)

			// Declare HTTP request method and URL
			req, err := http.NewRequest(requestMethod, targetURL.String(), requestBody)
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
			if followRedirects {
				client = http.Client{
					Jar: jar,
					Transport: &http.Transport{
						TLSClientConfig: &tls.Config{
							InsecureSkipVerify: true,
						},
					},
					Timeout: 20 * time.Second,
				}
			} else {
				client = http.Client{
					Jar: jar,
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
						if *flagVerbose {
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

	// Wait for the URLs to finish sending
	urlsInProgress.Wait()

	// VERBOSE
	if *flagVerbose {
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
	if *flagVerbose {
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
			for uResp := range uniqueResponses {
				// Read the unique response body
				uRespBody, err := readResponseBody(uResp)
				if err != nil {
					errors <- fmt.Errorf("Error reading unique response body: %s", err.Error())

					// Exit the inner loop
					continue
				}

				// Compare the response bodies
				respBodyMatch := false
				if string(respBody) == string(uRespBody) {
					respBodyMatch = true
				}

				// Compare response status code, body content, and content length
				if resp.StatusCode == uResp.StatusCode && resp.ContentLength == uResp.ContentLength && respBodyMatch {
					// Similar, increase count
					uniqueResponses[uResp]++
					// Exit inner loop
					continue
				} else {
					// Unique, add to unique responses
					uniqueResponses[resp] = 0
					// Exit inner loop
					continue
				}
			}
		}
	}

	// VERBOSE
	if *flagVerbose {
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

// TODO: Add option to send a second request at the same time, the same number of times (useful for adding 2 values to a database)
// TODO: Add option to include multiple session cookie values. Cookies for each request will be semicolon-delimited, and newline characters will delimit cookies for different requests.
