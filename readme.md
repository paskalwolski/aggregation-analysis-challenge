# The Solution

I chose to complete this challenge in Go. This was quite a challenge as Go is a completely new language to me - but it was an interesting exercise in learning functionality, techniques - and getting into the nitty-gritty. I hope that, while this took a bit longer to complete, it will be an invaluable learning experience.

At the core, there are two technical challenges here: creating a web server to accept and respond to HTTP requests, and listening to the SSE stream for post data.

## Web Server

Simple Web server (using the `net/http` package) that listens on port 8080.  
It expects only GET requests on the `analysis` endpoint, with two query parameters:

1. duration - of the format [amount][time identifier] (eg. 13s, 24h)
2. dimension - a string specifying the type of post data to search for.
   Functions dealing with the web server - and associated URL parsing - are placed directly in the root `aggregation-server` file. Utilities dealing with the incoming request are placed in the `utilities` file - allowing for separation of concerns.

### My Issues

- **Query Parsing**Originally I was going to use the `gorilla/mux` package (despite this being against the rules) but I found out that the `http` package gives access to URL parameters too - crisis avoided.
- **Error Handling** Handling errors in the query parsing was tricky - I am not used to the error handling patterns of Go. I settled on using an idiomatic `err` value where possible, and when one was encountered, writing an Error to the response body (using the explicit `http.Error` function) and returning the main Response. This gave some flexibility in being able to keep the 'flow' of the program quite obvious. I think this would have been avoided with the gorilla muxer, but alas.
  It also would have been good to be a little more explicit in the type of error found - JSON returns with error/message/details keys.

## SSE Stream

Streaming the body of the upfluence/stream response (currently usign a bufio scanner - could be converted to a direct JSON stream)
The timing is done by creating two goroutines - one to handle a timer, and one to handle the 'infinite' reading of this stream.
When the timer channel is fulfilled, the response body is closed and the resulting json WILL BE handled.

### My Issues

- **Using the Response** - I thought the body of a Response was just a JSON body, and was preparing to receive more 'messages' - but as it turns out, it is a stream which can be read by a scanner. As long as you are reading that stream, the connection remains open. Not sure what the timeout/response limit - is this using 'Keep-Alive' Header?
- **Resource Management** - Cleaning up the Timer and the Scanner if either case was interrupted was a challenge. Although in these cases the resources are quite minor, handling them effectively is of course extremely important.
- **The Event Stream** - The stream responses/messages did not conform too strictly to JSON rules (unquoted fields, no opening brace/floating closing brace) and that made it particularly tricky/frustrating to work with. This also made it impossible to use a JSON Decoder, and so I had to stick with the bufio scanner :/
