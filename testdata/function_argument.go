package main

func functionArgument(callback func()) {
	go func() {
		callback()
	}()
}
