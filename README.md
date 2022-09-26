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
* Install module via `go get`
```bash
$ go get github.com/mavjs/pushnotifier
```

This package also provides a rudimentary command line application called: `pnctl`. To install:
```bash
$ go get github.com/mavjs/pushnotifier/cmd/pnctl
```

## Usage
### Main package
```go
import (
    "github.com/mavjs/pushnotifier"
)

pn := pushnotifier.NewClient(nil, "dev.myapps.pn", "BBCCVV1122...", "ZZXX11ff...")
```
#### Sending Notification Messages
```go
// Sends a notification with text "hello world" to all registered devices silently.
// Meaning it will not create a notification sound/vibration.
pn.SendText("hello world", nil, true)

// Sends a notification with text "dlrow olleh" to selected devices.
// Meaning it will create a notification sound and or vibration.
pn.SendText("dlrow olleh", []string{"abcd", "efgh"}, false)

// Sends a notification with url "http://example.com"
pn.SendURL("http://example.com", []string{"abcd", "efgh"}, false)

// Sends a notification with text "hello world" and url "http://example.com"
// Upon clicking the notification this will open the default browser to the provided URL.
pn.SendNotification("hello world", "http://example.com", []string{"abcd", "efgh"}, false)

// Sends a notification with an images
pn.SendImage("path/to/image.png", []string{"abcd", "efgh"}, false)
```

#### Get Basic Information
```go
pn.GetDevices()

[GetDevices] Registered devices for user obtained
Devices: [abcd efgh ijkl]
```

#### Refresh App Token
```go
pn.RefreshToken()

[RefreshToken] App Token for user obtained
```

### Command Line Application - pnctl
```bash
$ pnctl help

pnctl - a commandline application that can be used to send different types of notifications to your registered devices.
You can send text and or url, or image as notifications.

Usage:
  pnctl [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  getdevices  Get connected devices
  help        Help about any command
  register    Registers API authentication details
  send        Sends different types of content to registered devices.

Flags:
      --config string   config file (default is /home/user/.config/pushnotifier/pushnotifier.yaml)
  -h, --help            help for pnctl

Use "pnctl [command] --help" for more information about a command.
```