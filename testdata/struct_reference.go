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
