// +build !debug

package main

// no exported profiler when not in debug mode
func initDebugInterfaces() {}
