package utils

import (
	"fmt"
	"log"
)

func Nlog(format string, args ...interface{}) {
	if DEBUG {
		log.Printf(format, args...)
	}
}

func Dlog(id string, format string, args ...interface{}) {
	if DEBUG {
		format = fmt.Sprintf("[%d] ", id) + format
		log.Printf(format, args...)
	}
}