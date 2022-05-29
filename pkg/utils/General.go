package utils

import (
	"fmt"
	"hash/fnv"
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

func Tlog(format string, args ...interface{}) {
	if TEST_LOG {
		log.Printf(format, args...)
	}
}

func Hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}