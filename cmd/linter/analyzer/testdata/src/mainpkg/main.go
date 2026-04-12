package main

import (
	"log"
	"os"
)

func main() {
	// log.Fatal and os.Exit are allowed here
	// because func main() of package main
	log.Fatal("bye")
	os.Exit(1)

	panic("oops") // want "usage of panic is not allowed"
}

func helper() {
	log.Fatal("bye") // want "usage of log.Fatal is not allowed"
	os.Exit(1)       // want "usage of os.Exit is not allowed"
	panic("oops")    // want "usage of panic is not allowed"
}
