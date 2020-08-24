# Retry

[![PkgGoDev](https://pkg.go.dev/badge/vthiery/retry)](https://pkg.go.dev/vthiery/retry)
![Build Status](https://github.com/vthiery/retry/workflows/Test/badge.svg)
![GolangCI Lint](https://github.com/vthiery/retry/workflows/GolangCI/badge.svg)

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
				100*time.Millisecond, // initialWait
				1*time.Second,        // maxWait
				2.0,                  // exponentFactor
				2*time.Millisecond,   // maximumJitterInterval
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
