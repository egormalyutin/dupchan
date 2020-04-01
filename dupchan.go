// This package provides duplicable channels, which can help you to organize
// your asynchronous application.
package dupchan

import "reflect"

// Original channel wrapped in reflect.Value
type originalChannel = reflect.Value

// Duplicate channel wrapped in reflect.Value
type duplicateChannel = reflect.Value

// Channel for communication with subscriber
type subscriberChannel = chan *duplicateChannel

// Global map needed to match original channel and its subscriber
var subscribers = map[originalChannel]subscriberChannel{}

// Global map needed to match duplicate channel and its subscriber. Used to
// send Unsubscribe message to subscriber
var unsubscribers = map[duplicateChannel]subscriberChannel{}

// Create duplicate channel, where all messages from original channel will
// appear. You can create multiple duplicate channels and all of them will have
// the same messages from original channel. After Duplicate was used, channel
// cannot be used in the same way as before, because copying goroutine is still
// working. You can kill it using `StopDuplication`.
//
// Actually, this function just starts a goroutine which copies all incoming
// messages into duplicated channels. Once message will be read, it won't be
// returned to original channel. Duplicated channels will close when
// original channel will close.
//
// You can pass additional argument to create buffered duplicate channel.
func Duplicate(o interface{}, b ...int) interface{} {
	orig := reflect.ValueOf(o)
	tp := orig.Type()

	if tp.Kind() != reflect.Chan {
		panic("argument must be a channel")
	}

	if _, e := subscribers[orig]; !e {
		subscribers[orig] = make(subscriberChannel)
		go subscribe(orig)
	}

	buff := 0
	if len(b) > 0 {
		buff = b[0]
	}

	dup := reflect.MakeChan(tp, buff)
	subscribers[orig] <- &dup
	unsubscribers[dup] = subscribers[orig]
	return dup.Interface()
}

// Unduplicate channel. This function sends message to copying goroutine to stop
// copying to this channel.
func Unduplicate(d interface{}) {
	dup := reflect.ValueOf(d)
	if dup.IsNil() {
		panic("argument must not be nil")
	}
	unsubscribers[dup] <- &dup
}

// Kill copying goroutine and stop duplication. Can be used on original or
// duplicated goroutine.
func StopDuplication(c interface{}) {
	ch := reflect.ValueOf(c)
	if ch.IsNil() {
		panic("argument must not be nil")
	}

	sub, ok := subscribers[ch]
	if ok {
		sub <- nil
		return
	}

	sub, ok = unsubscribers[ch]
	if ok {
		sub <- nil
	}
}

// Goroutine which copies incoming messages from original channel to
// duplicated channels.
func subscribe(orig originalChannel) {
	// Set with channel duplicates
	dups := map[duplicateChannel]bool{}

	cases := []reflect.SelectCase{
		// Received data from original channel
		{
			Dir:  reflect.SelectRecv,
			Chan: orig,
		},
		// Received new duplicate channel
		{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(subscribers[orig]),
		},
	}

	for {
		switch index, value, ok := reflect.Select(cases); index {
		// Received data from original channel
		case 0:
			// If original channel is closed, we close all duplicate
			// channels and remove subscriber channel
			if !ok {
				for dup := range dups {
					dup.Close()
				}
				delete(subscribers, orig)
				return
			}

			// Send data from original to duplicate channels
			for dup := range dups {
				dup.Send(value)
			}

		// Received new duplicate channel
		case 1:
			chn := value.Interface().(*duplicateChannel)
			// If nil is received, kill goroutine
			if chn == nil {
				return
			}

			ch := *chn
			// If duplicate channel not exists in duplicates set, then we
			// add it. Otherwise duplicate channel was received from
			// Unduplicate and we remove it
			if _, ok := dups[ch]; ok {
				delete(dups, ch)
			} else {
				dups[ch] = true
			}
		}
	}
}
