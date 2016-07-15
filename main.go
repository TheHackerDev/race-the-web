package main

import (
	"bufio"
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

// Responses is a channel to store the concurrent responses from the target
var responses chan *http.Response

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

// Number of requests
var numRequests int

// Request type
var requestType string

// Follow redirects
var followRedirects bool

// Verbose logging enabled
var verbose bool

// Command-line flags
var flagTargetURL = flag.String("url", "", "URL to send the request to.")
var flagBodyFile = flag.String("body", "", "The location (relative or absolute path) of a file containing the body of the request.")
var flagCookiesFile = flag.String("cookies", "", "The location (relative or absolute path) of a file containing newline-separate cookie values being sent along with the request. Cookie names and values are separated by a comma. For example: cookiename,cookieval")
var flagNumRequests = flag.Int("requests", 100, "The number of requests to send to the destination URL.")
var flagRequestType = flag.String("request", "POST", "The request type. Can be either `POST, GET, HEAD, PUT`.")
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
	errChan := sendRequests()
	if len(errChan) != 0 {
		for err := range errChan {
			log.Printf("[ERROR] %s\n", err.Error())
		}
	}

	// Compare the responses for uniqueness
	uniqueResponses := compareResponses()

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

	// Set verbose logging explicitely
	verbose = *flagVerbose

	// Determine whether to follow redirects
	followRedirects = *flagFollowRedirects

	// Set the request type
	switch strings.ToUpper(*flagRequestType) {
	case "POST":
		requestType = "POST"
	case "GET":
		requestType = "GET"
	case "PUT":
		requestType = "PUT"
	case "HEAD":
		requestType = "HEAD"
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

	// Check the number of requests used for testing
	numRequests = *flagNumRequests

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
func sendRequests() chan error {
	// Initialize the concurrency objects
	responses = make(chan *http.Response, numRequests)
	errorChannel := make(chan error, numRequests)
	urlsInProgress.Add(numRequests)

	// VERBOSE
	if verbose {
		log.Printf("[VERBOSE] Sending %d %s requests to %s\n", numRequests, requestType, targetURL.String())
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
			req, err := http.NewRequest(requestType, targetURL.String(), requestBody)
			if err != nil {
				errorChannel <- fmt.Errorf("Error in forming request: %v", err.Error())
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
						if verbose {
							log.Printf("[VERBOSE] %v\n", rErr)
						}
						// Add the response to the responses channel, because it is still valid
						responses <- resp
					} else {
						// URL Error, but not a redirect error
						errorChannel <- fmt.Errorf("Error in request #%v: %v\n", index, err)
					}
				} else {
					// Other type of error
					errorChannel <- fmt.Errorf("Error in request #%v: %v\n", index, err)
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
	if verbose {
		log.Printf("[VERBOSE] Requests complete.")
	}

	// Close the response and error chanels, so they don't block on the range read
	close(responses)
	close(errorChannel)

	return errorChannel
}

// Function compareResponses compares the responses returned from the requests,
// and adds them to a map, where the key is an *http.Response, and the value is
// the number of similar responses observed.
func compareResponses() (uniqueResponses map[*http.Response]int) {
	// Initialize the unique responses map
	uniqueResponses = make(map[*http.Response]int)

	// VERBOSE
	if verbose {
		log.Printf("[VERBOSE] Unique response comparison begin.\n")
	}

	// Compare the responses, one at a time
	for resp := range responses {
		// Add an entry, if the unique responses map is empty
		if len(uniqueResponses) == 0 {
			uniqueResponses[resp] = 0
		} else {
			// Add to the unique responses map, if no similar ones exist
			for uResp := range uniqueResponses {
				// Compare response status code, body content, and content length
				if resp.StatusCode == uResp.StatusCode && resp.ContentLength == uResp.ContentLength {
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
	if verbose {
		log.Printf("[VERBOSE] Unique response comparision complete.\n")
	}

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
		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("[Body] ")
		} else {
			fmt.Printf("[Body]\n%s\n", respBody)
			// Close the response body
			resp.Body.Close()
		}
		fmt.Printf("Similar: %v\n\n", count)
	}
}

// BUG: Not reading some response bodies. Might be a timeout issue?
// TODO: Compare response body as well (if content-length != 0)
