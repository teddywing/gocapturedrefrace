package main

import (
	"git.teddywing.com/gocapturedrefrace"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(gocapturedrefrace.Analyzer)
}
