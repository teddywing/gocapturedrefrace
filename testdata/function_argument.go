package main

func functionArgument(callback func()) {
	go func() {
		callback()
	}()
}

func functionVariable() {
	callback := func() int { return 0; }

	go func() {
		callback()
	}()
}
