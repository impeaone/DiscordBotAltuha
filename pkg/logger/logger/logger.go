package logger

import (
	"DiscordBotAltuha/pkg/Error"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type Log struct{}

func NewLog() *Log { return &Log{} }

func (logs *Log) Info(message string, place string) {
	log.Println("\nLevel: Info" + "\nMessage: " + message + "\nPlace: " + place + "\n")
	WriteLogsToFile(time.Now().Format("02.01.2006 15:04:05") +
		"\nLevel: Info" + "\nMessage: " + message + "\nPlace: " + place)
}

func (logs *Log) Warning(message string, place string) {
	log.Println("\nLevel: Warning" + "\nMessage: " + message + "\nPlace: " + place + "\n")
	WriteLogsToFile(time.Now().Format("02.01.2006 15:04:05") +
		"\nLevel: Warning" + "\nMessage: " + message + "\nPlace: " + place)
}

func (logs *Log) Error(message string, place string) {
	log.Println("\nLevel: Error" + "\nMessage: " + message + "\nPlace: " + place + "\n")
	WriteLogsToFile(time.Now().Format("02.01.2006 15:04:05") +
		"\nLevel: Error" + "\nMessage: " + message + "\nPlace: " + place)
}

func GetPlace() string {
	_, file, line, _ := runtime.Caller(1)
	split := strings.Split(file, "/")
	StartFile := split[len(split)-1]
	place := StartFile + ":" + strconv.Itoa(line)
	return place
}

func WriteLogsToFile(LogText string) {
	file, err := os.OpenFile(
		"../Log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println("\nLevel: Error" + "\nMessage: " + Error.LogFileDoesNotOpen + ": " + err.Error() + "\nPlace: " +
			GetPlace() + "\n")
	}
	_, err = file.WriteString(LogText + "\n\n")
	if err != nil {
		log.Println("\nLevel: Error" + "\nMessage: " + Error.LogFileDoesNotWrite + ": " + err.Error() + "\nPlace: " +
			GetPlace() + "\n")
	}
	file.Close()
}
