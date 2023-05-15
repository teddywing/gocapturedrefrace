package main

func main() {
	capturedReference := 0
	capturedReference2 := 1

	go func() {
		capturedReference += 1
		capturedReference2 += 1

		if capturedReference == 1 {
			return
		}
	}()
}
