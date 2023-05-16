// Copyright (c) 2023  Teddy Wing
//
// This file is part of Gocapturedrefrace.
//
// Gocapturedrefrace is free software: you can redistribute it and/or
// modify it under the terms of the GNU General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// Gocapturedrefrace is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU
// General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Gocapturedrefrace. If not, see
// <https://www.gnu.org/licenses/>.


package main

import "strings"

func main() {
	capturedReference := 0
	capturedReference2 := 1
	copied := 0

	go func(copied int) {
		capturedReference += 1 // want "captured reference capturedReference in goroutine closure"
		capturedReference2 += 1 // want "captured reference capturedReference2 in goroutine closure"
		copied += 1

		if capturedReference == 1 { // want "captured reference capturedReference in goroutine closure"
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

	go func(s *aStruct) { // want "reference s in goroutine closure"
		s.field += 1
	}(&s)
}
