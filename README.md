gocapturedrefrace
=================

[![GoDoc](https://godocs.io/gopkg.teddywing.com/capturedrefrace?status.svg)][Documentation]

An analyser that reports captured variable references in goroutine closures.

Goroutines that run function closures can capture reference variables from outer
scopes which could lead to data races. This analyzer checks closures run by
goroutines and reports uses of all variables declared in outer scopes, as well
as arguments to the closure with a pointer type.


## Example
Given the following program:

``` go
package main

func main() {}

type Record struct{}

func (r *Record) reticulateSplines() {}

type Spline struct {
	Curvature float64
}

func (r *Record) CapturedReference() {
	capturedReference := 0
	spline := &Spline{Curvature: 5.0}

	go func(s *Spline) {
		capturedReference += 1 // closure captures the variable
		// 'capturedReference' in a goroutine, which could
		// lead to data races

		if capturedReference > 0 {
			r.reticulateSplines() // goroutine closure captures 'r'
		}

		s.Curvature = 3.0 // 's' is a pointer type which could
		// lead to data races
	}(spline)
}
```

the analyser produces the following results:

	$ gocapturedrefrace ./...
	package_doc_example.go:17:10: reference s in goroutine closure
	package_doc_example.go:18:3: captured reference capturedReference in goroutine closure
	package_doc_example.go:22:6: captured reference capturedReference in goroutine closure
	package_doc_example.go:23:4: captured reference r in goroutine closure


## Install

	$ go install gopkg.teddywing.com/capturedrefrace/cmd/gocapturedrefrace@latest


## License
Copyright © 2023 Teddy Wing. Licensed under the GNU GPLv3+ (see the included
COPYING file).


[Documentation]: https://godocs.io/gopkg.teddywing.com/capturedrefrace
