# dupchan

[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/egormalyutin/dupchan)

This package implements duplicable channels in Go, which can help you to organize your asynchronous application.

## Example
```go
orig := make(chan int)

// Create three workers
for i := 0; i < 3; i++ {
	// and duplicate original channel in each of them.
	ch := dupchan.Duplicate(orig).(chan int)
	go func() {
		// Recieve and print message from duplicable channel.
		// If we wouldn't use duplicable channels, only one worker would
		// print the message, or we would need to send two messages
		// (for each worker)
		fmt.Println("->", <-ch)
	}()
}

ch2 := dupchan.Duplicate(orig).(chan int)
go func() {
	<-ch2
	panic("I won't panic")
	// Because ch2 will be unduplicated before message will be sent
}()
dupchan.Unduplicate(ch2)

orig <- 10
time.Sleep(time.Millisecond)
```

Output:
```
10
10
10
```
