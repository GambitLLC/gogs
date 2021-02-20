package logger

import "log"

func Printf(format string, v ...interface{}) {
	if len(v) == 0 {
		log.Printf("[INFO]: " + format)
	} else {
		log.Printf("[INFO]: "+format, v)
	}
}

func Error(v ...interface{}) {
	log.Fatalf("[ERROR]: %v\n", v)
}

func Errorf(format string, v ...interface{}) {
	log.Fatalf("[ERROR]: "+format, v)
}
