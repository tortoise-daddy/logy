package logy

import (
	"fmt"
	"os"
)

var handlers = []func(){}

func runHandler(handler func()) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Fprintln(os.Stderr, "Error: Logrus exit handler error:", err)
		}
	}()

	handler()
}

func runHandlers() {
	for _, handler := range handlers {
		runHandler(handler)
	}
}

// Exit runs all the Logrus atexit handlers and then terminates the program using os.Exit(code)
func Exit(code int) {
	runHandlers()
	os.Exit(code)
}
func RegisterExitHandler(handler func()) {
	handlers = append(handlers, handler)
}
func DeferExitHandler(handler func()) {
	handlers = append([]func(){handler}, handlers...)
}
