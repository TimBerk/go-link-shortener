// Package noosexit implements an analyzer that detects direct calls to os.Exit
// in the main function of the main package.
//
// Such calls can prevent proper cleanup and make code harder to test.
// Instead, it's recommended to:
//   - Return an error from main() and let the runtime handle exit
//   - Use a wrapper function that can be mocked in tests
//   - Structure your program to have a clear shutdown sequence
//
// Example of problematic code:
//
//	package main
//
//	import "os"
//
//	func main() {
//		os.Exit(1) // will be flagged by the analyzer
//	}
//
// Suggested alternatives:
//
//	package main
//
//	import (
//		"fmt"
//		"os"
//	)
//
//	func run() int {
//		// application logic here
//		return 1
//	}
//
//	func main() {
//		os.Exit(run())
//	}
package noosexit
