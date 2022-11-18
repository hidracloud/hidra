# http
HTTP plugin is used to make HTTP requests
## Available actions
### statusCodeShouldBe
Checks if the status code is equal to the expected value
#### Parameters
- statusCode: The expected status code
### addHTTPHeader
Adds a HTTP header to the request. If the header already exists, it will be overwritten
#### Parameters
- key: The header name
- value: The header value
### allowInsecureTLS
Allows insecure TLS connections. This is useful for testing purposes, but should not be used in production
#### Parameters
### forceIP
Forces the IP address to use for the request
#### Parameters
- ip: The IP address
### onFailure
Executes the steps if the previous step failed
#### Parameters
### onClose
Executes the steps when the test is finished
#### Parameters
### request
Makes a HTTP request
#### Parameters
-  (optional) method: The HTTP method
- url: The URL
-  (optional) body: The body
### bodyShouldContain
[DEPRECATED] Please use outputShouldContain from string plugin. Checks if the body contains the expected value
#### Parameters
- search: The expected value
### shouldRedirectTo
Checks if the response redirects to the expected URL
#### Parameters
- url: The expected URL
### setUserAgent
Sets the User-Agent header
#### Parameters
- user-agent: The User-Agent value
### followRedirects
Follows the redirect
#### Parameters
