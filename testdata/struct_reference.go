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

type AStruct struct {
	field string
}

func (s *AStruct) setField(value string) {
	s.field = value
}

func (s *AStruct) method2() {
	go func() {
		s.setField("test") // want "captured reference s in goroutine closure"
	}()
}
