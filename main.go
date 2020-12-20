package main

import (
	"context"
	"fmt"
	"time"
)

//Future methods
type Future interface {
	get() Result
	getWithTimeout(duration time.Duration) Result
	cancel() bool
	isCancelled() bool
	isRunning() bool
	complete() bool
}

// FutureTask Future object
type FutureTask struct {
	success        bool
	error          error
	running        bool
	done           bool
	result         Result
	channel        <-chan Result
	callbackMethod func()
}

// Result holds the result of callable
type Result struct {
	value interface{}
	error error
}

func (futureTask *FutureTask) get() Result {
	if futureTask.done {
		return futureTask.result
	}
	if futureTask.callbackMethod != nil {
		defer futureTask.callbackMethod()
	}
	ctx := context.Background()
	return futureTask.getWithContext(ctx)
}

func (futureTask *FutureTask) getWithTimeout(timeout time.Duration) Result {
	if futureTask.done {
		return futureTask.result
	}
	if futureTask.callbackMethod != nil {
		defer futureTask.callbackMethod()
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return futureTask.getWithContext(ctx)
}

func (futureTask *FutureTask) getWithContext(ctx context.Context) Result {
	fmt.Println("Executing getContext to receive from channel")
	select {
	case <-ctx.Done():
		futureTask.done = true
		futureTask.success = false
		futureTask.error = &TimeoutError{errorString: "Request Timeout!"}
		futureTask.result = Result{value: nil, error: futureTask.error}
		return futureTask.result

	case futureTask.result = <-futureTask.channel:
		if futureTask.result.error != nil {
			futureTask.done = true
			futureTask.success = false
			futureTask.error = futureTask.result.error
		} else {
			futureTask.success = true
			futureTask.done = true
			futureTask.error = nil
		}
		return futureTask.result
	}
}

func (futureTask *FutureTask) isCancelled() bool {
	if futureTask.done {
		if futureTask.error != nil && futureTask.error.Error() == "Cancelled Manually" {
			return true
		}
	}
	return false
}

func (futureTask *FutureTask) complete() bool {
	if futureTask.done {
		return true
	}
	return false
}

func (futureTask *FutureTask) cancel() bool {
	if futureTask.complete() || futureTask.isCancelled() || futureTask.isRunning() {
		return false
	}
	if futureTask.callbackMethod != nil {
		defer futureTask.callbackMethod()
	}
	interruptionError := &InterruptError{errorString: "Cancelled Manually"}
	futureTask.done = true
	futureTask.success = false
	futureTask.error = interruptionError
	futureTask.result = Result{value: nil, error: interruptionError}
	return true
}

func (futureTask *FutureTask) isRunning() bool {
	if futureTask.running {
		return true
	}
	return false
}

//Stringer method for result
func (result Result) String() string {
	err := "no"
	if result.error != nil {
		err = result.error.Error()
	}
	return fmt.Sprintf("%v with (%s error)", result.value, err)
}
func (futureTask *FutureTask) addDoneCallback(callbackMethod func()) {
	futureTask.callbackMethod = callbackMethod
}

//ReturnAFuture creates a new future for task func
func ReturnAFuture(task func() Result) *FutureTask {
	channelForExecution := make(chan Result)
	futureObject := FutureTask{
		success: false,
		done:    false,
		error:   nil,
		result:  Result{},
		channel: channelForExecution,
	}
	go func() {
		defer func() {
			close(channelForExecution)
			futureObject.running = false
		}()
		futureObject.running = true
		resultObject := task()
		channelForExecution <- resultObject
	}()
	return &futureObject
}

func (e *TimeoutError) Error() string {
	return e.errorString
}
func (e *InterruptError) Error() string {
	return e.errorString
}

//TimeoutError class
type TimeoutError struct {
	errorString string
}

//InterruptError class
type InterruptError struct {
	errorString string
}

func main() {
	//simple example of future
	futureInstance2 := ReturnAFuture(func() Result {
		var res interface{}
		res = "40"
		time.Sleep(4 * time.Second)
		return Result{value: res}
	})
	f2 := futureInstance2.get()
	fmt.Println(f2)
	fmt.Println("------------")

	//example of timeout error
	futureInstance1 := ReturnAFuture(func() Result {
		var res interface{}
		res = 30 + 23
		time.Sleep(2 * time.Second)
		return Result{value: res}
	})
	f1 := futureInstance1.getWithTimeout(1 * time.Second)
	fmt.Println(f1)
	fmt.Println("------------")

	//example of cancel operation
	futureInstance3 := ReturnAFuture(func() Result {
		var res interface{}
		res = "50"
		time.Sleep(20 * time.Second)
		return Result{value: res}
	})
	ok := futureInstance3.cancel()
	fmt.Println("Cancel operation:", ok)
	f3 := futureInstance3.get()
	fmt.Println(f3)
	fmt.Println("------------")

	//example of callback
	futureInstance4 := ReturnAFuture(func() Result {
		var res interface{}
		res = "50"
		time.Sleep(2 * time.Second)
		return Result{value: res}
	})
	futureInstance4.addDoneCallback(func() {
		fmt.Println("Executing callback function")
	})
	f4 := futureInstance4.get()
	fmt.Println(f4)
	fmt.Println("------------")

}
