[![Go Report Card](https://goreportcard.com/badge/github.com/aaronhnatiw/race-the-web)](https://goreportcard.com/report/github.com/aaronhnatiw/race-the-web) [![Build Status](https://travis-ci.org/aaronhnatiw/race-the-web.svg?branch=master)](https://travis-ci.org/aaronhnatiw/race-the-web)

# Race The Web (RTW)

Tests for race conditions in web applications by sending out a user-specified number of requests to a target URL (or URLs) *simultaneously*, and then compares the responses from the server for uniqueness. Includes a number of configuration options.

## UPDATE: Now CI Compatible!

Version 2.0.0 now makes it easier than ever to integrate RTW into your continuous integration pipeline (à la [Jenkins](https://jenkins.io/), [Travis](https://travis-ci.org/), or [Drone](https://github.com/drone/drone)), through the use of an easy to use HTTP API. More information can be found in the **Usage** section below.

## Watch The Talk

[![Racing the Web - Hackfest 2016](https://img.youtube.com/vi/4T99v957I0o/0.jpg)](https://www.youtube.com/watch?v=4T99v957I0o)

_Racing the Web - Hackfest 2016_

Slides: https://www.slideshare.net/AaronHnatiw/racing-the-web-hackfest-2016

## Usage

With configuration file

```sh
$ race-the-web config.toml
```

API

```sh
$ race-the-web
```

### Configuration File

**Example configuration file included (_config.toml_):**

```toml
# Sample Configurations

# Send 100 requests to each target
count = 100
# Enable verbose logging
verbose = true
# Use an http proxy for all connections
proxy = "http://127.0.0.1:8080"

# Specify the first request
[[requests]]
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

# Specify the second request
[[requests]]
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

### API

Since version 2.0.0, RTW now has a full-featured API, which allows you to easily integrate it into your continuous integration (CI) tool of choice. This means that you can quickly and easily test your web application for race conditions automatically whenever you commit your code.

The API works through a simple set of HTTP calls. You provide input in the form of JSON and receive a response in JSON. The 3 API endpoints are as follows:

- `POST` `http://127.0.0.1:8000/set/config`: Provide configuration data (in JSON format) for the race condition test you want to run (examples below).
- `GET` `http://127.0.0.1:8000/get/config`: Fetch the current configuration data. Data is returned in a JSON response.
- `POST` `http://127.0.0.1:8000/start`: Begin the race condition test using the configuration that you have already provided. All findings are returned back in JSON output.

#### Example JSON configuration (sent to `/set/config` using a `POST` request)

```json
{
    "count": 100,
    "verbose": false,
    "requests": [
        {
            "method": "POST",
            "url": "http://racetheweb.io/bank/withdraw",
            "cookies": [
                "sessionId=dutwJx8kyyfXkt9tZbboT150TjZoFuEZGRy8Mtfpfe7g7UTPybCZX6lgdRkeOjQA"
            ],
            "body": "amount=1",
            "redirects": true
        }
    ]
}
```

#### Example workflow using curl


1. Send the configuration data

```sh
$ curl -d '{"count":100,"verbose":false,"requests":[{"method":"POST","url":"http://racetheweb.io/bank/withdraw","cookies":["sessionId=Ay2jnxL2TvMnBD2ZF-5bXTXFEldIIBCpcS4FLB-5xjEbDaVnLbf0pPME8DIuNa7-"],"body":"amount=1","redirects":true}]}' -H "Content-Type: application/json" -X POST http://127.0.0.1:8000/set/config

{"message":"configuration saved"}
```

2. Retrieve the configuration data for validation

```sh
$ curl -X GET http://127.0.0.1:8000/get/config

{"count":100,"verbose":false,"proxy":"","requests":[{"method":"POST","url":"http://racetheweb.io/bank/withdraw","body":"amount=1","cookies":["sessionId=Ay2jnxL2TvMnBD2ZF-5bXTXFEldIIBCpcS4FLB-5xjEbDaVnLbf0pPME8DIuNa7-"],"headers":null,"redirects":true}]}
```

3. Start the race condition test

```sh
$ curl -X POST http://127.0.0.1:8000/start
```

Response (expanded for visibility):

```JSON
[
    {
        "Response": {
            "Body": "\n<!DOCTYPE html>\n<html lang=\"en\">\n  <head>\n    <meta charset=\"utf-8\">\n    <meta http-equiv=\"X-UA-Compatible\" content=\"IE=edge\">\n    <meta name=\"viewport\" content=\"width=device-width, initial-scale=1\">\n    \n    <title>Bank Test</title>\n\n    \n    <link href=\"/static/css/bootstrap.min.css\" rel=\"stylesheet\">\n\n    \n    \n    \n\n    \n    <meta name=\"twitter:card\" content=\"summary_large_image\" />\n    <meta name=\"twitter:site\" content=\"@insp3ctre\" />\n    <meta name=\"twitter:title\" content=\"Race Condition Exploit Practice\" />\n    <meta name=\"twitter:description\" content=\"Learn how to exploit race conditions in web applications.\" />\n    <meta name=\"twitter:image\" content=\"/static/img/bank_homepage_screenshot_wide.png\" />\n    <meta name=\"twitter:image:alt\" content=\"Image of the bank account exploit application.\" />\n  </head>\n  <body>\n    <nav class=\"navbar\">\n      <div class=\"container-fluid\">\n        <div class=\"navbar-header\">\n          <a class=\"navbar-brand\" href=\"/\">Race-The-Web</a>\n        </div>\n        <ul class=\"nav navbar-nav\">\n          <li><a href=\"/bank\">Bank</a></li>\n        </ul>\n        <ul class=\"nav navbar-nav navbar-right\">\n          <li><a href=\"https://www.youtube.com/watch?v=4T99v957I0o\"><img src=\"http://racetheweb.io/static/img/logo-youtube.png\" alt=\"Racing the Web - Hackfest 2016\" title=\"Racing the Web - Hackfest 2016\"></a></li>\n          <li><a href=\"https://github.com/insp3ctre/race-the-web\"><img src=\"/static/img/logo-github.png\" alt=\"Race-The-Web on Github\"></a></li>\n        </ul>\n      </div>\n    </nav>\n\n    <div class=\"container\">\n        <div class=\"row\">\n            <div class=\"page-header\">\n                <h1 class=\"text-center\">Welcome to SpeedBank, International</h1>\n            </div>\n        </div>\n        \n        <div class=\"row\">\n            <div class=\"col-xs-12 col-sm-8 col-sm-offset-2\">\n                <p class=\"text-center bg-success\">You have successfully withdrawn $1</p>\n            </div>\n        </div>\n        \n        \n        <div class=\"row\">\n            <h2 class=\"text-center\">Balance: 9999</h2>\n        </div>\n        <div class=\"row\">\n            <div class=\"col-xs-8 col-xs-offset-3\">\n                <form action=\"/bank/withdraw\" method=\"POST\" class=\"form-inline\">\n                    <div class=\"form-group\">\n                        <label class=\"sr-only\" for=\"withdrawAmount\">Amount (in dollars)</label>\n                        <div class=\"input-group\">\n                            <div class=\"input-group-addon\">$</div>\n                            <input type=\"text\" class=\"form-control\" id=\"withdrawAmount\" name=\"amount\" placeholder=\"Amount\">\n                            <div class=\"input-group-addon\">.00</div>\n                        </div>\n                        <div class=\"input-group\">\n                            <input type=\"submit\" class=\"btn btn-primary\" value=\"Withdraw cash\">\n                        </div>\n                    </div>\n                </form>\n            </div>\n        </div>\n        \n        <div class=\"row\">\n            <div class=\"col-xs-12 col-sm-8 col-sm-offset-2\">\n                <h2 class=\"text-center\">Instructions</h2>\n                <ol>\n                    <li>Click “Initialize” to initialize a bank account with $10,000.</li>\n                    <li>Withdraw money from your account, observe that your account balance is updated, and that you have received the amount requested.</li>\n                    <li>Repeat the request with <a href=\"https://github.com/insp3ctre/race-the-web\">race-the-web</a>. Your config file should look like the following:</li>\n<pre>\n# Make one request\ncount = 100\nverbose = true\n[[requests]]\n    method = \"POST\"\n    url = \"http://racetheweb.io/bank/withdraw\"\n    # Withdraw 1 dollar\n    body = \"amount=1\"\n    # Insert your sessionId cookie below.\n    cookies = [“sessionId=&lt;insert here&gt;\"]\n    redirects = false\n</pre>\n                    <li>Visit the bank page again in your browser to view your updated balance. Note that the total <em>should</em> be $100 less ($1 * 100 requests) than when you originally withdrew money. However, due to a race condition flaw in the application, your balance will be much more, yet you will have received the money from the bank in every withdrawal.</li>\n                </ol>\n            </div>\n        </div>\n    </div>\n    \n    <script type=\"text/javascript\">\n        \n        history.replaceState(\"Bank\", \"Bank\", \"/bank\")\n    </script>\n    \n\n    <p class=\"small text-center\">\n        <span class=\"glyphicon glyphicon-copyright-mark\" aria-hidden=\"true\"></span><a href=\"https://www.twitter.com/insp3ctre\">Aaron Hnatiw</a> 2017\n    </p>\n    \n    <script src=\"https://ajax.googleapis.com/ajax/libs/jquery/1.12.4/jquery.min.js\"></script>\n    \n    <script src=\"/static/js/bootstrap.min.js\"></script>\n    \n    <script>\n    (function(i,s,o,g,r,a,m){i['GoogleAnalyticsObject']=r;i[r]=i[r]||function(){\n    (i[r].q=i[r].q||[]).push(arguments)},i[r].l=1*new Date();a=s.createElement(o),\n    m=s.getElementsByTagName(o)[0];a.async=1;a.src=g;m.parentNode.insertBefore(a,m)\n    })(window,document,'script','https://www.google-analytics.com/analytics.js','ga');\n\n    ga('create', 'UA-93555669-1', 'auto');\n    ga('send', 'pageview');\n\n    </script>\n    </body>\n</html>\n",
            "StatusCode": 200,
            "Length": -1,
            "Protocol": "HTTP/1.1",
            "Headers": {
                "Content-Type": [
                    "text/html; charset=utf-8"
                ],
                "Date": [
                    "Fri, 18 Aug 2017 15:36:29 GMT"
                ]
            },
            "Location": ""
        },
        "Targets": [
            {
                "method": "POST",
                "url": "http://racetheweb.io/bank/withdraw",
                "body": "amount=1",
                "cookies": [
                    "sessionId=Ay2jnxL2TvMnBD2ZF-5bXTXFEldIIBCpcS4FLB-5xjEbDaVnLbf0pPME8DIuNa7-"
                ],
                "headers": null,
                "redirects": true
            }
        ],
        "Count": 1
    },
    {
        "Response": {
            "Body": "\n<!DOCTYPE html>\n<html lang=\"en\">\n  <head>\n    <meta charset=\"utf-8\">\n    <meta http-equiv=\"X-UA-Compatible\" content=\"IE=edge\">\n    <meta name=\"viewport\" content=\"width=device-width, initial-scale=1\">\n    \n    <title>Bank Test</title>\n\n    \n    <link href=\"/static/css/bootstrap.min.css\" rel=\"stylesheet\">\n\n    \n    \n    \n\n    \n    <meta name=\"twitter:card\" content=\"summary_large_image\" />\n    <meta name=\"twitter:site\" content=\"@insp3ctre\" />\n    <meta name=\"twitter:title\" content=\"Race Condition Exploit Practice\" />\n    <meta name=\"twitter:description\" content=\"Learn how to exploit race conditions in web applications.\" />\n    <meta name=\"twitter:image\" content=\"/static/img/bank_homepage_screenshot_wide.png\" />\n    <meta name=\"twitter:image:alt\" content=\"Image of the bank account exploit application.\" />\n  </head>\n  <body>\n    <nav class=\"navbar\">\n      <div class=\"container-fluid\">\n        <div class=\"navbar-header\">\n          <a class=\"navbar-brand\" href=\"/\">Race-The-Web</a>\n        </div>\n        <ul class=\"nav navbar-nav\">\n          <li><a href=\"/bank\">Bank</a></li>\n        </ul>\n        <ul class=\"nav navbar-nav navbar-right\">\n          <li><a href=\"https://www.youtube.com/watch?v=4T99v957I0o\"><img src=\"http://racetheweb.io/static/img/logo-youtube.png\" alt=\"Racing the Web - Hackfest 2016\" title=\"Racing the Web - Hackfest 2016\"></a></li>\n          <li><a href=\"https://github.com/insp3ctre/race-the-web\"><img src=\"/static/img/logo-github.png\" alt=\"Race-The-Web on Github\"></a></li>\n        </ul>\n      </div>\n    </nav>\n\n    <div class=\"container\">\n        <div class=\"row\">\n            <div class=\"page-header\">\n                <h1 class=\"text-center\">Welcome to SpeedBank, International</h1>\n            </div>\n        </div>\n        \n        <div class=\"row\">\n            <div class=\"col-xs-12 col-sm-8 col-sm-offset-2\">\n                <p class=\"text-center bg-success\">You have successfully withdrawn $1</p>\n            </div>\n        </div>\n        \n        \n        <div class=\"row\">\n            <h2 class=\"text-center\">Balance: 9998</h2>\n        </div>\n        <div class=\"row\">\n            <div class=\"col-xs-8 col-xs-offset-3\">\n                <form action=\"/bank/withdraw\" method=\"POST\" class=\"form-inline\">\n                    <div class=\"form-group\">\n                        <label class=\"sr-only\" for=\"withdrawAmount\">Amount (in dollars)</label>\n                        <div class=\"input-group\">\n                            <div class=\"input-group-addon\">$</div>\n                            <input type=\"text\" class=\"form-control\" id=\"withdrawAmount\" name=\"amount\" placeholder=\"Amount\">\n                            <div class=\"input-group-addon\">.00</div>\n                        </div>\n                        <div class=\"input-group\">\n                            <input type=\"submit\" class=\"btn btn-primary\" value=\"Withdraw cash\">\n                        </div>\n                    </div>\n                </form>\n            </div>\n        </div>\n        \n        <div class=\"row\">\n            <div class=\"col-xs-12 col-sm-8 col-sm-offset-2\">\n                <h2 class=\"text-center\">Instructions</h2>\n                <ol>\n                    <li>Click “Initialize” to initialize a bank account with $10,000.</li>\n                    <li>Withdraw money from your account, observe that your account balance is updated, and that you have received the amount requested.</li>\n                    <li>Repeat the request with <a href=\"https://github.com/insp3ctre/race-the-web\">race-the-web</a>. Your config file should look like the following:</li>\n<pre>\n# Make one request\ncount = 100\nverbose = true\n[[requests]]\n    method = \"POST\"\n    url = \"http://racetheweb.io/bank/withdraw\"\n    # Withdraw 1 dollar\n    body = \"amount=1\"\n    # Insert your sessionId cookie below.\n    cookies = [“sessionId=&lt;insert here&gt;\"]\n    redirects = false\n</pre>\n                    <li>Visit the bank page again in your browser to view your updated balance. Note that the total <em>should</em> be $100 less ($1 * 100 requests) than when you originally withdrew money. However, due to a race condition flaw in the application, your balance will be much more, yet you will have received the money from the bank in every withdrawal.</li>\n                </ol>\n            </div>\n        </div>\n    </div>\n    \n    <script type=\"text/javascript\">\n        \n        history.replaceState(\"Bank\", \"Bank\", \"/bank\")\n    </script>\n    \n\n    <p class=\"small text-center\">\n        <span class=\"glyphicon glyphicon-copyright-mark\" aria-hidden=\"true\"></span><a href=\"https://www.twitter.com/insp3ctre\">Aaron Hnatiw</a> 2017\n    </p>\n    \n    <script src=\"https://ajax.googleapis.com/ajax/libs/jquery/1.12.4/jquery.min.js\"></script>\n    \n    <script src=\"/static/js/bootstrap.min.js\"></script>\n    \n    <script>\n    (function(i,s,o,g,r,a,m){i['GoogleAnalyticsObject']=r;i[r]=i[r]||function(){\n    (i[r].q=i[r].q||[]).push(arguments)},i[r].l=1*new Date();a=s.createElement(o),\n    m=s.getElementsByTagName(o)[0];a.async=1;a.src=g;m.parentNode.insertBefore(a,m)\n    })(window,document,'script','https://www.google-analytics.com/analytics.js','ga');\n\n    ga('create', 'UA-93555669-1', 'auto');\n    ga('send', 'pageview');\n\n    </script>\n    </body>\n</html>\n",
            "StatusCode": 200,
            "Length": -1,
            "Protocol": "HTTP/1.1",
            "Headers": {
                "Content-Type": [
                    "text/html; charset=utf-8"
                ],
                "Date": [
                    "Fri, 18 Aug 2017 15:36:30 GMT"
                ]
            },
            "Location": ""
        },
        "Targets": [
            {
                "method": "POST",
                "url": "http://racetheweb.io/bank/withdraw",
                "body": "amount=1",
                "cookies": [
                    "sessionId=Ay2jnxL2TvMnBD2ZF-5bXTXFEldIIBCpcS4FLB-5xjEbDaVnLbf0pPME8DIuNa7-"
                ],
                "headers": null,
                "redirects": true
            }
        ],
        "Count": 1
    },
    {
        "Response": {
            "Body": "\n<!DOCTYPE html>\n<html lang=\"en\">\n  <head>\n    <meta charset=\"utf-8\">\n    <meta http-equiv=\"X-UA-Compatible\" content=\"IE=edge\">\n    <meta name=\"viewport\" content=\"width=device-width, initial-scale=1\">\n    \n    <title>Bank Test</title>\n\n    \n    <link href=\"/static/css/bootstrap.min.css\" rel=\"stylesheet\">\n\n    \n    \n    \n\n    \n    <meta name=\"twitter:card\" content=\"summary_large_image\" />\n    <meta name=\"twitter:site\" content=\"@insp3ctre\" />\n    <meta name=\"twitter:title\" content=\"Race Condition Exploit Practice\" />\n    <meta name=\"twitter:description\" content=\"Learn how to exploit race conditions in web applications.\" />\n    <meta name=\"twitter:image\" content=\"/static/img/bank_homepage_screenshot_wide.png\" />\n    <meta name=\"twitter:image:alt\" content=\"Image of the bank account exploit application.\" />\n  </head>\n  <body>\n    <nav class=\"navbar\">\n      <div class=\"container-fluid\">\n        <div class=\"navbar-header\">\n          <a class=\"navbar-brand\" href=\"/\">Race-The-Web</a>\n        </div>\n        <ul class=\"nav navbar-nav\">\n          <li><a href=\"/bank\">Bank</a></li>\n        </ul>\n        <ul class=\"nav navbar-nav navbar-right\">\n          <li><a href=\"https://www.youtube.com/watch?v=4T99v957I0o\"><img src=\"http://racetheweb.io/static/img/logo-youtube.png\" alt=\"Racing the Web - Hackfest 2016\" title=\"Racing the Web - Hackfest 2016\"></a></li>\n          <li><a href=\"https://github.com/insp3ctre/race-the-web\"><img src=\"/static/img/logo-github.png\" alt=\"Race-The-Web on Github\"></a></li>\n        </ul>\n      </div>\n    </nav>\n\n    <div class=\"container\">\n        <div class=\"row\">\n            <div class=\"page-header\">\n                <h1 class=\"text-center\">Welcome to SpeedBank, International</h1>\n            </div>\n        </div>\n        \n        <div class=\"row\">\n            <div class=\"col-xs-12 col-sm-8 col-sm-offset-2\">\n                <p class=\"text-center bg-success\">You have successfully withdrawn $1</p>\n            </div>\n        </div>\n        \n        \n        <div class=\"row\">\n            <h2 class=\"text-center\">Balance: 9997</h2>\n        </div>\n        <div class=\"row\">\n            <div class=\"col-xs-8 col-xs-offset-3\">\n                <form action=\"/bank/withdraw\" method=\"POST\" class=\"form-inline\">\n                    <div class=\"form-group\">\n                        <label class=\"sr-only\" for=\"withdrawAmount\">Amount (in dollars)</label>\n                        <div class=\"input-group\">\n                            <div class=\"input-group-addon\">$</div>\n                            <input type=\"text\" class=\"form-control\" id=\"withdrawAmount\" name=\"amount\" placeholder=\"Amount\">\n                            <div class=\"input-group-addon\">.00</div>\n                        </div>\n                        <div class=\"input-group\">\n                            <input type=\"submit\" class=\"btn btn-primary\" value=\"Withdraw cash\">\n                        </div>\n                    </div>\n                </form>\n            </div>\n        </div>\n        \n        <div class=\"row\">\n            <div class=\"col-xs-12 col-sm-8 col-sm-offset-2\">\n                <h2 class=\"text-center\">Instructions</h2>\n                <ol>\n                    <li>Click “Initialize” to initialize a bank account with $10,000.</li>\n                    <li>Withdraw money from your account, observe that your account balance is updated, and that you have received the amount requested.</li>\n                    <li>Repeat the request with <a href=\"https://github.com/insp3ctre/race-the-web\">race-the-web</a>. Your config file should look like the following:</li>\n<pre>\n# Make one request\ncount = 100\nverbose = true\n[[requests]]\n    method = \"POST\"\n    url = \"http://racetheweb.io/bank/withdraw\"\n    # Withdraw 1 dollar\n    body = \"amount=1\"\n    # Insert your sessionId cookie below.\n    cookies = [“sessionId=&lt;insert here&gt;\"]\n    redirects = false\n</pre>\n                    <li>Visit the bank page again in your browser to view your updated balance. Note that the total <em>should</em> be $100 less ($1 * 100 requests) than when you originally withdrew money. However, due to a race condition flaw in the application, your balance will be much more, yet you will have received the money from the bank in every withdrawal.</li>\n                </ol>\n            </div>\n        </div>\n    </div>\n    \n    <script type=\"text/javascript\">\n        \n        history.replaceState(\"Bank\", \"Bank\", \"/bank\")\n    </script>\n    \n\n    <p class=\"small text-center\">\n        <span class=\"glyphicon glyphicon-copyright-mark\" aria-hidden=\"true\"></span><a href=\"https://www.twitter.com/insp3ctre\">Aaron Hnatiw</a> 2017\n    </p>\n    \n    <script src=\"https://ajax.googleapis.com/ajax/libs/jquery/1.12.4/jquery.min.js\"></script>\n    \n    <script src=\"/static/js/bootstrap.min.js\"></script>\n    \n    <script>\n    (function(i,s,o,g,r,a,m){i['GoogleAnalyticsObject']=r;i[r]=i[r]||function(){\n    (i[r].q=i[r].q||[]).push(arguments)},i[r].l=1*new Date();a=s.createElement(o),\n    m=s.getElementsByTagName(o)[0];a.async=1;a.src=g;m.parentNode.insertBefore(a,m)\n    })(window,document,'script','https://www.google-analytics.com/analytics.js','ga');\n\n    ga('create', 'UA-93555669-1', 'auto');\n    ga('send', 'pageview');\n\n    </script>\n    </body>\n</html>\n",
            "StatusCode": 200,
            "Length": -1,
            "Protocol": "HTTP/1.1",
            "Headers": {
                "Content-Type": [
                    "text/html; charset=utf-8"
                ],
                "Date": [
                    "Fri, 18 Aug 2017 15:36:36 GMT"
                ]
            },
            "Location": ""
        },
        "Targets": [
            {
                "method": "POST",
                "url": "http://racetheweb.io/bank/withdraw",
                "body": "amount=1",
                "cookies": [
                    "sessionId=Ay2jnxL2TvMnBD2ZF-5bXTXFEldIIBCpcS4FLB-5xjEbDaVnLbf0pPME8DIuNa7-"
                ],
                "headers": null,
                "redirects": true
            }
        ],
        "Count": 98
    }
]
```

## Binaries

The program has been written in Go, and as such can be compiled to all the common platforms in use today. The following architectures have been compiled, and can be found in the [releases](https://github.com/insp3ctre/race-the-web/releases) tab:

- Windows amd64
- Windows 386
- Linux amd64
- Linux 386
- OSX amd64
- OSX 386

## Compiling

First, make sure you have Go installed on your system. If you don't you can follow the install instructions for your operating system of choice here: https://golang.org/doc/install.

Build a binary for your current CPU architecture

```sh
$ make build
```

Build for all major CPU architectures (see [Makefile](https://github.com/insp3ctre/race-the-web/blob/master/Makefile) for details) at once

```sh
$ make
```

### Dep

This project uses [Dep](https://github.com/golang/dep) for dependency management. All of the required files are kept in the `vendor` directory, however if you are getting errors related to dependencies, simply download Dep

```sh
$ go get -u github.com/golang/dep/cmd/dep
```

and run the following command from the RTW directory in order to download all dependencies

```sh
$ dep ensure
```

### Go 1.7 and newer are supported

Before 1.7, the `encoding/json` package's `Encoder` did not have a method to escape the `&`, `<`, and `>` characters; this is required in order to have a clean output of full HTML pages when running these race tests. _If this is an issue for your test cases, please submit a [new issue](https://github.com/insp3ctre/race-the-web/issues) indicating as such, and I will add a workaround (just note that any output from a server with those characters will come back with unicode escapes instead)._ Here are the relevant release details from Go 1.7: https://golang.org/doc/go1.7#encoding_json.

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

- Concurrency: Concurrency primitives are built into the language itself, and extremely easy to add to any Go program. Threading is [handled by the Go runtime scheduler](https://morsmachine.dk/go-scheduler), and not by the underlying operating system, which allows for some serious performance optimizations when it comes to concurrency.
- Compiled: *Cross-compiles* to [most modern operating systems](https://golang.org/doc/install/source#environment); not slowed down by an interpreter or virtual machine middle-layer ([here are some benchmarks vs Java](https://benchmarksgame.alioth.debian.org/u64q/go.html)). (Oh, and did I mention that the binaries are statically compiled?)
- Lightweight: Only [25 keywords](https://golang.org/ref/spec#Keywords) in the language, and yet still almost everything can be done using the standard library.

For more of the nitty-gritty details on why Go is so fast, see [Dave Cheney](https://twitter.com/davecheney)'s [excellent talk on the subject](http://dave.cheney.net/2014/06/07/five-things-that-make-go-fast), from 2014.
