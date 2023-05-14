package main

func main() {
	capturedReference := 0

	go func() {
		capturedReference += 1
	}()
}
