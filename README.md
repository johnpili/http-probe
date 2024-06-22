# HTTP Probe

## Overview
Use this tool to display and analyze HTTP headers and HTTP body payloads. 
It is primarily used for testing HTTP clients, microservices and 
other HTTP based applications. I developed this tool to help my development 
team analyzing HTTP messages coming from a black box systems during early 
stage of development.

## Usage
Use the browser to access the application's landing page. It will generate
a unique ID that you will use to send messages.

## Isolated Sessions
This tool can handle multiple users who want to send and analyze HTTP messages. 
The unique session ID's isolate users from each other

## Disclaimer
No information sent into this tool are stored. 
It uses an event based mechanism (using SocketIO) to display the HTTP 
information into the user interface (UI) in real-time. A cookie is used to 
generate and established a unique session ID per user.
