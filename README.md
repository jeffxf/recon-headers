# recon-headers
A web server that replies to requests with a unique png image and logs the details of every request. 

## Purpose
There are existing services that provide the capability of tracking when an email you have sent has been opened. One of which is the service MailTracker. These services are great for what they do but aren't really aimed towards recon purposes and can't be self-hosted.

## How To Use

Start the web server and come up with URI's to send out. The server will respond with a randomly generated png image to any URI requested. You can take advantage of the wildcard by providing each recipient with a different URI path that identifies each individual. Then wait for requests to the image URI which will be logged to the logfile.

It's up to your creativity as to how you send the URL in a way that it will be opened by the recipient. Examples could be an email or word document with an image attached using the URL. When the email or word doc is opened, their host will make a request to the URL to get the image. The generated images are almost completely transparent and relatively small so the recipient shouldn't notice it.

## Features

- Can reply back to wildcard URI's
- Replies with randomly generated images
- Attempts to prevent caching
- Logs:
    - Date/time
    - Source IP
    - Source port
    - Returned status code
    - Request URI
    - Request headers
- Only uses standard go packages

### Default Settings
    - ip: listens on all interfaces
    - port: 8080
    - uri: /*
    - logfile: recon-headers.log

### Help
```
$ ./recon-headers -h
Usage of recon-headers:
  -ip string
        The local IP address the web server should listen on (default "All interfaces")
  -logfile string
        The name of the log file (default "recon-headers.log")
  -port string
        The port number the web server should listen on (default "8080")
  -uri string
        The URI that returns an image
        Examples:
                "/"             =       Respond to any path or file name (wildcard path and file name)
                "/recon"        =       Only respond to exact match of "/recon"
                "/recon.png"    =       Only respond to exact match of "/recon.png"
                "/recon/"       =       Respond to "/recon/" and anything after (wildcard file name)
        (default "/")
```

### Running Example
```
$ ./recon-headers -port 3333 -logfile engagement01.log -uri "/eng01/"
[*] Starting web server (:3333)
```

## Binaries
The 3 binaries in this repo are for x86 Windows, OSX, and Linux. They were built using Docker following the "Building Your Own" section below. You of course only need one of the binaries for whichever OS you'll be running this server on.

### File hashes:
```
# md5sum recon-headers.*
8a38facdd29fe080cd1ad0828eaf7d6f  recon-headers.linux
dfcc70328e9dce4938dac1f43e31ef06  recon-headers.osx
1eab99fd79cf0de7a9b8dd3467b5e7d0  recon-headers.win
```

## Building Your Own
Building the binary yourself is easy using the go compiler:

```
go build -v -o recon-headers recon-headers.go
```

Or, if you prefer to use Docker to build, the following will compile the source for x86 Windows, OSX, and Linux:

```
$ docker run --rm -it -v "$PWD":/usr/src/recon-headers -w /usr/src/recon-headers golang:1.10 bash

# GOOS=linux GOARCH=386 go build -v -ldflags "-s -w" -o recon-headers.linux recon-headers.go

# GOOS=darwin GOARCH=386 go build -v -ldflags "-s -w" -o recon-headers.osx recon-headers.go

# GOOS=windows GOARCH=386 go build -v -ldflags "-s -w" -o recon-headers.win recon-headers.go
```

## Why Golang?

I've been using go a lot recently and enjoy it so here we are. I understand you probably would rather this be written using an interpreted lanuage like python. A few reasons could be:

- Easier to modify a python script and quickly run it without having to install go and recompile
    - I agree and might write this in python as well if I get around to it since it shouldn't take long at all. (But also, building in docker instead of installing go locally is super easy...)
- Downloading a python script and running it is less sketchy than a binary
    - Meh.. just compile it yourself then (again, you should be using docker these days anyways)
