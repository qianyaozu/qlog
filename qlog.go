package qlog

import (
	"fmt"
	"log"
	"os"
	"time"
)

type LogStruct struct {
	Title string
	Level string
	Data  interface{}
}

var LogChannel = make(chan LogStruct, 10000)
var pathArray = make([]string, 0)

func init() {
	fmt.Println("start logging")
	go HandlerLogs()
}
func HandlerLogs() {
	for log := range LogChannel {
		SaveLog(log.Level, log.Title, log.Data)
	}
}
func SaveLog(level, title string, data interface{}) {
	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
		}
	}()
	dir := "log/" + time.Now().Format("20060102") + "/" + level + "/"
	for _, p := range pathArray {
		if p == dir {
			goto saveFile
		}
	}

	if len(pathArray) > 100 {
		//清理过多的缓存
		pathArray = make([]string, 0)
	}
	if _, err := os.Stat(dir); err == nil {
		err = os.MkdirAll(dir, 0777)
		if err != nil {
			fmt.Println("SaveLog", err)
			return
		} else {
			pathArray = append(pathArray, dir)
		}
	} else {
		fmt.Println("SaveLog", err)
	}

saveFile:
	fileName := dir + title + ".log"
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("open file error !", err)
		return
	}
	defer file.Close()
	logs := log.New(file, "", log.LstdFlags)
	if data != nil {
		logs.Println(data)
	}else{
		fmt.Println("empty",level, title , data )
	}
}

func Error(title string, data interface{}) {
	Log("Error", title, data)
}
func Warn(title string, data interface{}) {
	Log("Warn", title, data)
}
func Info(title string, data interface{}) {
	Log("Info", title, data)
}
func Trace(title string, data interface{}) {
	Log("Trace", title, data)
}
func Fatal(title string, data interface{}) {
	Log("Fatal", title, data)
}

func Log(level, title string, data interface{}) {
	var log LogStruct
	log.Title = title
	log.Level = level
	log.Data = data
	LogChannel <- log
}
