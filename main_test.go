package main

import (
	"fmt"
	"runtime"
	"testing"
)

func TestMain(t *testing.T) {
	fmt.Printf("cpu: %d\n", runtime.NumCPU())
}
