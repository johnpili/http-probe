# HTTP Probe

## Overview                        

Use this tool to display HTTP headers and HTTP body payloads. It is primarily used
for testing HTTP clients, microservices and other HTTP event base messages. 
I developed this tool to help my development team in mocking up a SOA server which
we don't have access during development.

![ScreenShot](https://raw.github.com/johnpili/http-probe/master/http-probe-demo.gif)

## Usage
                        
Using your web browser or Postman do an HTTP GET or HTTP POST request to the generated URL
e.g.: https://<span></span>probe.johnpili.com/send/{generated-id}
                                                     
## Monitor Isolation
                        
This tool can handle multiple users who want to debug their HTTP messages. To make that 
possible, the tool generates a unique ID for you to send and monitor you HTTP transactions.            

## Disclaimer                        

No information sent into this tool are stored. It uses an event based mechanism (websocket) to 
display the HTTP information into the user interface (UI) in real-time. A cookie is used to 
generate and established an ID reference to the web UI monitor.                        