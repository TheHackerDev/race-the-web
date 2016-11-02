# Race The Web

Tests for race conditions in web applications by sending out a user-specified number of requests to a target URL (or URLs) *simultaneously*, and then compares the responses from the server for uniqueness. Includes a number of configuration options.

## Usage

`race-the-web config.toml`

### Configuration File

**Example configuration file included (_config.toml_):**

```toml
# Sample Configurations

# Send 100 requests to each target
count = 100
# Enable verbose logging
verbose = true

# Specify the first target
[[target]]
    # Use the GET request method
    method = "GET"
    # Set the URL target. Any valid URL is accepted, including ports, https, and parameters.
    url = "https://example.com/pay?val=1000"
    # Set the request body.
    # body = "body=text"
    # Set the cookie values to send with the request to this target. Must be an array.
    cookies = ["PHPSESSIONID=12345","JSESSIONID=67890"]
    # Set custom headers to send with the request to this target. Must be an array.
    headers = ["X-Originating-IP: 127.0.0.1", "X-Remote-IP: 127.0.0.1"]
    # Follow redirects
    redirects = true

# Specify the second target
[[target]]
    # Use the POST request method
    method = "POST"
    # Set the URL target. Any valid URL is accepted, including ports, https, and parameters.
    url = "https://example.com/pay"
    # Set the request body.
    body = "val=1000"
    # Set the cookie values to send with the request to this target. Must be an array.
    cookies = ["PHPSESSIONID=ABCDE","JSESSIONID=FGHIJ"]
    # Set custom headers to send with the request to this target. Must be an array.
    headers = ["X-Originating-IP: 127.0.0.1", "X-Remote-IP: 127.0.0.1"]
    # Do not follow redirects
    redirects = false
```

TOML Spec: https://github.com/toml-lang/toml

## Binaries

The program has been written in Go, and as such can be compiled to all the common platforms in use today. The following architectures have been compiled, and can be found in the [releases](https://github.com/insp3ctre/race-the-web/releases) tab:

- Windows amd64
- Windows 386
- Linux amd64
- Linux 386
- OSX amd64
- OSX 386

Alternatively, you can compile the code yourself. See [Dave Cheney](https://twitter.com/davecheney)'s excellent [post](http://dave.cheney.net/2015/08/22/cross-compilation-with-go-1-5 "Cross-compilation with Go 1.5") on the topic.

## The Vulnerability

> A race condition is a flaw that produces an unexpected result when the timing of actions impact other actions. An example may be seen on a multithreaded application where actions are being performed on the same data. Race conditions, by their very nature, are difficult to test for.
> - [OWASP](https://www.owasp.org/index.php/Testing_for_Race_Conditions_(OWASP-AT-010))

Race conditions are a well known issue in software development, especially when you deal with fast, multi-threaded languages.

However, as network speeds get faster and faster, web applications are becoming increasingly vulnerable to race conditions. Often because of legacy code that was not created to handle hundreds or thousands of simultaneous requests for the same function or resource.

The problem can often only be discovered when a fast, multi-threaded language is being used to generate these requests, using a fast network connection; at which point it becomes a network and logic race between the client application and the server application.

That is where **Race The Web** comes in. This program aims to discover race conditions in web applications by sending a large amount of requests to a specific endpoint at the same time. By doing so, it may invoke unintended behaviour on the server, such as the duplication of user information, coupon codes, bitcoins, etc.

**Warning:** Denial of service may be an unintended side-effect of using this application, so please be careful when using it, and always perform this kind of testing with the explicit permission of the server owner and web application owner.

Credit goes to [Josip Franjković](https://twitter.com/josipfranjkovic) for his [excellent article on the subject](https://www.josipfranjkovic.com/blog/race-conditions-on-web), which introduced me to this problem.

## Why Go

The [Go programming language](https://golang.org/) is perfectly suited for the task, mainly because it is *so damned fast*. Here are a few reasons why:

* Concurrency: Concurrency primitives are built into the language itself, and extremely easy to add to any Go program. Threading is [handled by the Go runtime scheduler](https://morsmachine.dk/go-scheduler), and not by the underlying operating system, which allows for some serious performance optimizations when it comes to concurrency.
* Compiled: *Cross-compiles* to [most modern operating systems](https://golang.org/doc/install/source#environment); not slowed down by an interpreter or virtual machine middle-layer ([here are some benchmarks vs Java](https://benchmarksgame.alioth.debian.org/u64q/go.html)). (Oh, and did I mention that the binaries are statically compiled?)
* Lightweight: Only [25 keywords](https://golang.org/ref/spec#Keywords) in the language, and yet still almost everything can be done using the standard library.

For more of the nitty-gritty details on why Go is so fast, see [Dave Cheney](https://twitter.com/davecheney)'s [excellent talk on the subject](http://dave.cheney.net/2014/06/07/five-things-that-make-go-fast), from 2014.