package utils

import (
	"log"
)

var Log *log.Logger = log.New(log.Writer(), "[LOG]: ", log.Ltime|log.Lshortfile)
