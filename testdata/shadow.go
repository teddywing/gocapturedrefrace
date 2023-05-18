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

	go func() {
		// err is redeclared here and shadows the outer scope. No diagnostic
		// should be printed.
		var err error
		err = errors.New("shadowing declaration err")
		if err != nil {
			log.Print(err)
		}
	}()
}

func multiIdentifierDeclaration() {
	var err1, err2 error
	err1 = nil
	err2 = nil
	if err1 != nil || err2 != nil {
		log.Print(err1, err2)
	}

	go func() {
		// err1 and err2 are redeclared here and shadow the outer scope. No
		// diagnostic should be printed.
		var err1, err2 error
		err1 = errors.New("shadowing declaration err1")
		err2 = errors.New("shadowing declaration err2")
		if err1 != nil || err2 != nil {
			log.Print(err1, err2)
		}
	}()
}
