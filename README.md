# race-the-web

Tests race conditions in web applications by sending out a user-specified number of requests to a target URL *simultaneously*.

## The Vulnerability

> A race condition or race hazard is the behavior of an electronic, software or other system where the output is dependent on the sequence or timing of other uncontrollable events. It becomes a bug when events do not happen in the order the programmer intended. The term originates with the idea of two signals racing each other to influence the output first.
> - [Wikipedia] (https://en.wikipedia.org/wiki/Race_condition)

Race conditions are a well known issue in software development, especially when you deal with fast, multi-threaded languages.

However, as network speeds get faster and faster, web applications are becoming increasingly vulnerable to race conditions. Often because of legacy code that was not created to handle hundreds or thousands of simultaneous requests for the same function or resource.

The problem can often only be discovered when a fast, multi-threaded language is being used to generate these requests, using a fast network connection; at which point it becomes a network and logic race between the client application and the server application.

That is where **race-the-web** comes in. This program aims to discover race conditions in web applications by sending a large amount of requests to a specific endpoint at the same time. By doing so, it may invoke unintended behaviour on the server, such as the duplication of user information, coupon codes, bitcoins, etc.

**Warning:** Denial of service may be an unintended side-effect of using this application, so please be careful when using it, and always perform this kind of testing with the explicit permission of the server owner and web application owner.

Credit goes to [Josip Franjković](https://twitter.com/josipfranjkovic) for his [excellent article on the subject][https://www.josipfranjkovic.com/blog/race-conditions-on-web), which introduced me to this problem.

## Practical Examples

TODO: Use the program in the wild to gather practical examples of its effectiveness and the prevelance of race conditions in web applications.

## Usage

This is a command-line tool. Use the following flags to run the program:

- `-url`: The URL to send the request to.
- `-body`: The location (relative or absolute path) of a file containing the body of the request.
- `-cookies`: The location (relative or absolute path) of a file containing newline-separate cookie values being sent along with the request. Cookie names and values are separated by a comma. For example: `cookiename,cookieval`.
- `-requests`: The number of requests to send to the destination URL. (default: `100`)
- `-type`: The request type. Can be either `POST`, `GET`, `HEAD`, or `PUT`. (default: `POST`)
- `-redirects`: Follow redirects (`3xx` status code in responses).
- `-v`: Enable verbose logging to the console.

**Example:**

- `race-the-web -url=http://www.example.com/ -body=body.txt -cookies=cookies.txt -requests=100 -type=POST -v`
..* Sends 100 POST requests to `http://www.example.com/`, with the cookies found in the file `cookies.txt` and the body contents found in the file `body.txt` (both files being located in the same directory as the program is being run from). Verbose logging is also enabled.

## Why Go

The [Go programming language] (https://golang.org/) is perfectly suited for the task, mainly because it is *so damned fast*. Here are a few reasons why:

* Concurrency: Concurrency primitives are built into the language itself, and extremely easy to add to any Go program. Threading is [handled by the Go runtime scheduler](https://morsmachine.dk/go-scheduler), and not by the underlying operating system, which allows for some serious performance optimizations when it comes to concurrency.
* Compiled: *Cross-compiles* to [most modern operating systems](https://golang.org/doc/install/source#environment); not slowed down by an interpreter or virtual machine middle-layer ([here are some benchmarks vs Java](https://benchmarksgame.alioth.debian.org/u64q/go.html)). (Oh, and did I mention that the binaries are statically compiled?)
* Lightweight: Only [25 keywords](https://golang.org/ref/spec#Keywords) in the language, and yet still almost everything can be done using the standard library.

For more of the nitty-gritty details on why Go is so fast, see [Dave Cheney](https://twitter.com/davecheney)'s [excellent talk on the subject](http://dave.cheney.net/2014/06/07/five-things-that-make-go-fast), from 2014.
