package logger

import "log"

// TODO: consider using https://github.com/Sirupsen/logrus

func Printf(format string, v ...interface{}) {
	log.Printf("[INFO]: "+format, v...)
}

func Error(v ...interface{}) {
	log.Fatalf("[ERROR]: %v\n", v)
}

func Errorf(format string, v ...interface{}) {
	log.Fatalf("[ERROR]: "+format, v...)
}
