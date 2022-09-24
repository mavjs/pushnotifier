# PushNotifier (V2) for Go
A go module to easily interact with the service of [PushNotifier](https://pushnotifier.de/) in your go projects.

You can find the documentation on [pkg.go.dev](https://pkg.go.dev/github.com/mavjs/pushnotifier) 

## About
Using this go module you can easily send:
* messages
* URLs
* images

For more information visit [pushnotifier.de](https://pushnotifier.de/)

## Installation
* Install via `go get`
```bash
$ go get github.com/mavjs/pushnotifier
```

## Usage
```go
import (
    "github.com/mavjs/pushnotifier"
)

pn := pushnotifier.NewClient(nil, "dev.myapps.pn", "BBCCVV1122...", "ZZXX11ff...")
pn.SendText("hello world", false)
```