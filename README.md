# EventBus in Go

[![Tests](https://github.com/ksysoev/sebus/actions/workflows/main.yml/badge.svg)](https://github.com/ksysoev/sebus/actions/workflows/main.yml)
[![codecov](https://codecov.io/gh/ksysoev/sebus/graph/badge.svg?token=JKC3N0BOUU)](https://codecov.io/gh/ksysoev/sebus)
[![Go Report Card](https://goreportcard.com/badge/github.com/ksysoev/sebus)](https://goreportcard.com/report/github.com/ksysoev/sebus)
[![Go Reference](https://pkg.go.dev/badge/github.com/ksysoev/sebus.svg)](https://pkg.go.dev/github.com/ksysoev/sebus)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)



This repository contains an implementation of an EventBus in Go. An EventBus is a design pattern that offers a simple communication system between components in an application.

## Features

- Supports multiple subscribers: Multiple components can subscribe to the same event.
- Asynchronous event delivery: Events are delivered to subscribers asynchronously.

## Installation

To use EventBus in your Go project, you can install it using go get:

```
go get github.com/ksysoev/sebus
```


## Usage

First, create a new EventBus:

```go

import (
    "github.com/ksysoev/sebus"
)

eb := sebus.NewEventBus()
```

To publish an event, your event classes should satisfy the interface `sebus.Event`:

```go
event := YourEvent{...}
err := eb.Publish(event)
```

To subscribe to events:

```go
sub, err := eb.Subscribe("topic", 5)
```

To unsubscribe:

```go
err := eb.Unsubscribe(sub)
```

## License

This project is licensed under the MIT License - see the LICENSE file for details.