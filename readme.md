# The Solution

I chose to complete this challenge in Go. This was quite a challenge as Go is a completely new language to me - but it was an interesting exercise in learning functionality, techniques - and getting into the nitty-gritty. I hope that, while this took a bit longer to complete, it will be an invaluable learning experience.

At the core, there are three technical challenges here: 
  1. creating a web server to accept and respond to HTTP requests
  2. listening to the SSE stream for post data
  3. Extracting the data from the JSON stream. 

### Dependencies
This project only uses the Golang standard library
The packages used are:
- `net/http` - for dealing with HTTP requests (both incoming and outgoing)
- `net/url` - for extracting URL queries
- `encoding/json` - for interacting with JSON-formatted data
- `bufio` - for reading a data stream from the SSE Response
- `fmt` - for general string formatting
- `log` - for general logging
- `time` - for timing the response, as well as performance measurement 


## Web Server

Simple Web server (using the `net/http` package) that listens on port 8080.  
It expects only GET requests on the `analysis` endpoint, with two query parameters:

1. duration - of the format [amount][time identifier] (eg. 13s, 24h)
2. dimension - a string specifying the type of post data to search for.
   Functions dealing with the web server - and associated URL parsing - are placed directly in the root `aggregation-server` file. Worker Scripts and their utils dealing with the incoming request are placed in the `utilities` file (although I should have named this file `workers`) - allowing for separation of concerns.

There is no default value for either parameter - they must both be passed as URL queries. In the case that a parameter is missing, an error response is created with the error as plain text. This is definitely something that can be improved - returning a definied JSON error would be good. 


### My Issues

- **Query Parsing**Originally I was going to use the `gorilla/mux` package (despite this being against the rules) but I found out that the `http` package gives access to URL parameters too - crisis avoided.
- **Error Handling** Handling errors in the query parsing was tricky - I am not used to the error handling patterns of Go. I settled on using an idiomatic `err` value where possible, and when one was encountered, writing an Error to the response body (using the explicit `http.Error` function) and returning the main Response. This gave some flexibility in being able to keep the 'flow' of the program quite obvious. I think this would have been avoided with the gorilla muxer, but alas.
  It also would have been good to be a little more explicit in the type of error found - JSON returns with error/message/details keys.

## SSE Stream

A request is made to the SSE Stream endpoint (https://stream.upfluence.co/stream) which streams an assortment of post data. The stream is listened to for a set period of time (defined by the 'duration' URL Query) and then analysed. 
The timing is handled with the `time.After` function running on a seperate thread. While there is time remaining, the stream is scanned. When the time is over, the last scanned value from the stream is handled, and then the scanning loop ends.
 
I decided to keep this time limited to actually listening to the stream (and not including the time taken to connect there, which could take longer than the specified minimum interval) This introduced some issues as the request is 'listened to' for longer than the specified time (usually on the order of 0.2 seconds longer). I believe that this is because of the duration of the `scanner.Scan` function, as well as the time taken to complete a read from the scan and check the timing channel. 
Similarly, the time taken to analyse a single post's data could slow down the scanner, and so limit the amount of posts which can be analysed within the time period. This could be avoided by offloading the analysis to a different thread, allowing the scanner to continue scanning while posts are analysed. However, from testing this slowdown is negligible - on the order of 0.0002s per post. 

The request to the SSE stream also takes some time to handle (~1.5 seconds) This, combined with the time taken for pre- and post-request analysis, means that the endpoint tends to run for `n+2` seconds, where `n` is the duration specified in the URL query in seconds. 

### My Issues

- **Using the Response** - I thought the body of a Response was just a JSON body, and was preparing to receive more 'messages' - but as it turns out, it is a stream which can be read by a scanner. As long as you are reading that stream, the connection remains open. Not sure what the timeout/response limit - is this using 'Keep-Alive' Header?
- **Resource Management** - Cleaning up the Timer and the Scanner if either case was interrupted was a challenge. Although in these cases the resources are quite minor, handling them effectively is of course extremely important.


## Extracting Values 
Once the scanner reads a line from the data stream, it is analysed to extract a specific key. All the data stream lines are either blank lines, which are discarded, or JSON-formatted text conforming to the following: 
`data: {DATA_KEY: {key_1: value, key_2: value, ...}}`
The specified key from the `dimension` URL query is extracted from the second-inner object. The 'data' text is discareded, and the remaining information is martialled into a map, so that it can be (relatively) easily accessed. 
The amount of posts analysed is tracked, and the cumulative sum of the values is also tracked. Where the specified dimension does not exist, the post counter is incremented but the cumulative sum is unaffected. This may not be entirely correct, but can be easily changed. These counters are used to calculate an average(mean) value for that dimension across ALL posts within the time period. 
Timestamps are also extracted from each data point
These values are used to create the endpoint response, which takes the shape: 
```
{
  total_posts:        int
  minimum_timestamp:  int64
  maximuim_timstamp:  int64
  average_dimension:  float64
  
}
```
This is slightly different to the example response in the following ways: 
1. Instead of returning `average_likes` (when the endpoint is queried with likes) there is now a generic `average_dimension` key. This makes the data more easily interpreted on the other side - if necessary, a `dimension` key could be added to the response for analysis on the other side. 
2. The average(mean) value is returned as a float64. This extra precision can be trimmed if unnecessary, but ensures that the data is as precise as possible, and this API is not responsible for any loss of precision. 

The final response is created by the worker function, and handed back to the server script for encoding. This means that making the  average calculation concurrent using multiple threads will be more difficult, but that is a clear performance gain for a larger project - and mostly unnecessary in this case. As-is, the fact that we are calculating these values on the fly means there is already very little performance overhead. 

## My Issues
- **Static Typing** - I battled a lot with the unkown shape of the incoming JSON object. After extracting the `data` value, I knew it was of type `map[string]any` - but I was not able to effectively extract an integer value from that. Ended up having to jump through quite a few hoops of waiting for internal typing, and then type asserting and catching the errors associated. This was offloaded to the `extractNumericKey` function
- **The Event Stream** - The stream responses/messages did not conform too strictly to JSON rules (unquoted fields, no opening brace/floating closing brace) and that made it particularly tricky/frustrating to work with. This also made it impossible to use a JSON Decoder, and so I had to stick with the bufio scanner :/


# Testing the Application
Right now, there is no explicit testing available for this project. 
To be honest, this is because I haven't read anything about testing frameworks in Go - it is definitely the next area for me to explore. 

# Running the Application
The server can be run locally by running: 
```
cd aggregation-analysis-challenge
go run .
```

From there, the server will be hosted by default on `localhost:8080` and can be interacted with directly from a browser, or a service such as Postman. 