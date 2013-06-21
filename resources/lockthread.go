package resources

import (
	"runtime"
)

var funcChannel chan func()

func init() {
	funcChannel = make(chan func())

	go func() {
		// GOMAXPROCS(0) changes nothing
		runtime.GOMAXPROCS(runtime.GOMAXPROCS(0) + 1)
		runtime.LockOSThread()
		for {
			fn := <-funcChannel
			fn()
		}
	}()
}

// All functions passed to this function will be run in the same hardware
// thread. The RunInThread function will block untill the other function
// has finished.
func RunInThread(fn func()) {
	done := make(chan bool)
	funcChannel <- func() {
		fn()
		done <- true
	}
	_ = <-done
}

// Async version of RunInThread
func RunInThreadNB(fn func()) {
	funcChannel <- fn
}
