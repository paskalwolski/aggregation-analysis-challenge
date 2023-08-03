# The Solution
I chose to complete this challenge in Go. This was quite a challenge as Go is a completely new language to me - but it was an interesting exercise in learning functionality, techniques - and getting into the nitty-gritty. I hope that, while this took a bit longer to complete, it will be an invaluable learning experience. 

At the core, there are two technical challenges here: creating a web server to accept and respond to HTTP requests, and listening to the SSE stream for post data. 

## Web Server
Simple Web server (using the `net/http` package) that listens on port 8080.  

## My Issues
- Originally I was going to use the `gorilla/mux` package (despite this being against the rules) but I found out that the `http` package gives access to URL parameters too - crisis avoided.
- Handling errors in the query parsing was tricky - I am not used to the error handling patterns of Go. I settled on using a idiomatic `err` value where possible, and when one was encountered, writing an Error to the response body (using the explicit `http.Error` function) and returning the endpoint handler. This gave some flexibility in being able to keep the 'flow' of the program quite obvious. I think this would have been avoided with the gorilla muxer, but alas. 
- 


Usability testing was conducted using Postman


