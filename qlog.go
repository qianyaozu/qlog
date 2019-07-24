package qlog

import (
	"fmt"
	"github.com/gin-gonic/gin/json"
	"github.com/go-redis/redis"
	"log"
	"os"
	"syscall"
	"time"
)

type LogStruct struct {
	Title string
	Level string
	Data  interface{}
}

var LogChannel = make(chan LogStruct, 10000)
var pathArray = make([]string, 0)
var client = redis.NewClient(&redis.Options{
	Addr:         "192.168.2.207:6379",
	DialTimeout:  10 * time.Second,
	ReadTimeout:  30 * time.Second,
	WriteTimeout: 30 * time.Second,
	PoolSize:     10,
	PoolTimeout:  30 * time.Second,
})

func init() {
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
	if level == "ELK" {
		b, _ := json.Marshal(data)
		client.RPush("elk_log_queue", string(b)).Result()
		return
	}

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
	if _, err := os.Stat(dir); err != nil {
		err = os.MkdirAll(dir, 0777)
		if err != nil {
			fmt.Println("SaveLog 1", err)
			return
		} else {
			pathArray = append(pathArray, dir)
		}
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
	} else {
		fmt.Println("empty", level, title, data)
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
func ELK(title string, data interface{}) {
	Log("ELK", title, data)
}

func Log(level, title string, data interface{}) {
	var log LogStruct
	log.Title = title
	log.Level = level
	log.Data = data
	LogChannel <- log
}

func Panic(f *os.File) {
	redirectStderr(f)
}

var (
	kernel32         = syscall.MustLoadDLL("kernel32.dll")
	procSetStdHandle = kernel32.MustFindProc("SetStdHandle")
)

func setStdHandle(stdhandle int32, handle syscall.Handle) error {
	r0, _, e1 := syscall.Syscall(procSetStdHandle.Addr(), 2, uintptr(stdhandle), uintptr(handle), 0)
	if r0 == 0 {
		if e1 != 0 {
			return error(e1)
		}
		return syscall.EINVAL
	}
	return nil
}

// redirectStderr to the file passed in
func redirectStderr(f *os.File) {
	err := setStdHandle(syscall.STD_ERROR_HANDLE, syscall.Handle(f.Fd()))
	if err != nil {
		log.Fatalf("Failed to redirect stderr to file: %v", err)
	}
}
