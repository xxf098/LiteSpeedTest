package log

import (
	"fmt"
	"log"
	"os"
)

var (
	logger    = log.New(os.Stdout, "", log.Ldate|log.Ltime)
	loggerErr = log.New(os.Stderr, "", log.Ldate|log.Ltime)
	level     = ERROR
)

func SetLevel(newLevel LogLevel) {
	level = newLevel
}

type Message interface {
	String() string
}

// Write a simple logger
func Write(msg Message) {
	logger.Print(msg.String() + "\n")
}

func D(format string, v ...interface{}) {
	print(format, v...)
}

func Debug(format Message, v ...interface{}) {
	D(format.String(), v...)
}

func I(format string, v ...interface{}) {
	if INFO < level {
		return
	}
	print(format, v...)
}

func W(format string, v ...interface{}) {
	if WARNING < level {
		return
	}
	print(format, v...)
}

func Warnln(format string, v ...interface{}) {
	W(format, v)
}

func E(format string, v ...interface{}) {
	printErr(format, v...)
}

func Error(format Message, v ...interface{}) {
	E(format.String(), v...)
}

func print(msg string, args ...interface{}) {
	// m := fmt.Sprintf(msg, args...)
	for _, arg := range args {
		msg += fmt.Sprintf(" %v", arg)
	}
	logger.Println(msg)
}

func printErr(msg string, args ...interface{}) {
	for _, arg := range args {
		msg += fmt.Sprintf(" %v", arg)
	}
	loggerErr.Println(msg)
}
