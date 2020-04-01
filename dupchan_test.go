package dupchan_test

import (
	"fmt"
	"time"

	"github.com/egormalyutin/dupchan"
)

func Example() {
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
	// Output:
	// -> 10
	// -> 10
	// -> 10
}

func ExampleDuplicate() {
	orig := make(chan int)
	counter := 0

	for i := 0; i < 3; i++ {
		ch := dupchan.Duplicate(orig).(chan int)
		go func() {
			counter += <-ch
		}()
	}

	orig <- 3
	time.Sleep(time.Millisecond)
	fmt.Println(counter)
	// Output: 9
}

func ExampleUnduplicate() {
	orig := make(chan int)
	counter := 0

	ch := dupchan.Duplicate(orig).(chan int)
	go func() {
		counter += <-ch
	}()
	dupchan.Unduplicate(ch)

	orig <- 3
	time.Sleep(time.Millisecond)
	fmt.Println(counter)
	// Output: 0
}

func ExampleStopDuplication() {
	orig := make(chan int)
	counter := 0

	for i := 0; i < 3; i++ {
		ch := dupchan.Duplicate(orig).(chan int)
		go func() {
			counter += <-ch
		}()
	}

	orig <- 3

	dupchan.StopDuplication(orig)

	go func() {
		orig <- 3
	}()

	time.Sleep(time.Millisecond)
	fmt.Println(counter)
	// Output: 9
}
