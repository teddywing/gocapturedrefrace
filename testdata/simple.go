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

		var decl string
		decl = "b"
		strings.Repeat(decl, 2)
	}(copied)
}

func argumentReference() {
	type aStruct struct{
		field int
	}

	s := aStruct{field: 0}

	go func(s *aStruct) {
		s.field += 1
	}(&s)
}
