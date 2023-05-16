package main

import (
	"golang.org/x/tools/go/analysis/singlechecker"
	"gopkg.teddywing.com/gocapturedrefrace"
)

func main() {
	singlechecker.Main(gocapturedrefrace.Analyzer)
}
