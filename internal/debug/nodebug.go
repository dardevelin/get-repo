//go:build !debug

package debug

// No-op implementations for non-debug builds

func Log(format string, args ...interface{}) {
	// No-op
}

func LogFunction(name string) func() {
	// No-op
	return func() {}
}

func LogError(err error, context string) {
	// No-op
}

func LogState(component, oldState, newState string) {
	// No-op
}
