# Retry

[![PkgGoDev](https://pkg.go.dev/badge/vthiery/retry)](https://pkg.go.dev/github.com/vthiery/retry)
[![GitHub go.mod Go version of a Go module](https://img.shields.io/github/go-mod/go-version/vthiery/retry.svg)](https://github.com/vthiery/retry)
[![GoReportCard example](https://goreportcard.com/badge/github.com/vthiery/retry)](https://goreportcard.com/report/github.com/vthiery/retry)
![Build Status](https://github.com/vthiery/retry/workflows/Test/badge.svg)
![GolangCI Lint](https://github.com/vthiery/retry/workflows/GolangCI/badge.svg)
![License](https://img.shields.io/github/license/vthiery/retry)

## Description

Yet another retrier \o/

## Installation

```sh
go get -u github.com/vthiery/retry
```

## Usage

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/vthiery/retry"
)

func main() {
	// Define the retry strategy, with 10 attempts and an exponential backoff
	retry := retry.New(
		retry.WithMaxAttempts(10),
		retry.WithBackoff(
			retry.NewExponentialBackoff(
				100*time.Millisecond, // minWait
				1*time.Second,        // maxWait
				2*time.Millisecond,   // maxJitter
			),
		),
	)

	// A cancellable context can be used to stop earlier
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Define the function that can be retried
	fn := func() error {
		fmt.Println("doing something...")
		return errors.New("actually, can't do it 🤦")
	}

	// Call the `retry.Do` to attempt to perform `fn`
	if err := retry.Do(ctx, fn); err != nil {
		fmt.Printf("failed to perform `fn`: %v\n", err)
	}
}
```
