package otherpkg

import (
	"log"
	"os"
)

func DoSomething() {
	log.Fatal("bye") // want "usage of log.Fatal is not allowed"
	os.Exit(1)       // want "usage of os.Exit is not allowed"
	panic("oops")    // want "usage of panic is not allowed"
}
