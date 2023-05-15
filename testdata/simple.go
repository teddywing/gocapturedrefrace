package main

import "strings"

func main() {
	capturedReference := 0
	capturedReference2 := 1
	copied := 0

	go func(copied int) {
		capturedReference += 1
		capturedReference2 += 1
		copied += 1

		if capturedReference == 1 {
			return
		}

		newVar := 0
		newVar += 1

		str := "a"
		strings.Repeat(str, 3)
	}(copied)
}
