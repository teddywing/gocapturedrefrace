package main

import (
	"errors"
	"log"
)

func shadow() {
	var err error
	err = nil
	if err != nil {
		log.Print(err)
	}

	go func() {
		// err is redeclared here and shadows the outer scope. No diagnostic
		// should be printed.
		err := errors.New("shadowing err")
		if err != nil {
			log.Print(err)
		}
	}()
}
