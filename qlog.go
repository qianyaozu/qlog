package qlog

import (
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	"log"
	"os"
	"sync"
	"time"
)

type Options struct {
	Name         string
	RedisAddress string
	ChannelSize  int
}

var ioMap sync.Map

type QLog struct {
	Name         string
	redisClient  *redis.Client
	redisAddress string
	logChannel   chan LogStruct
	run          bool
}

type LogStruct struct {
	Title string
	Level string
	Data  interface{}
}

//构造日志对象
func NewQLog(option *Options) *QLog {
	qlog := &QLog{
		Name:         option.Name,
		run:          true,
		redisAddress: option.RedisAddress,
	}
	if qlog.Name == "" {
		qlog.Name = "qgate"
	}
	if option.RedisAddress != "" {
		qlog.redisClient = redis.NewClient(&redis.Options{
			Addr:         option.RedisAddress,
			DialTimeout:  10 * time.Second,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
			PoolSize:     10,
			PoolTimeout:  30 * time.Second,
		})
	}
	if option.ChannelSize <= 0 {
		option.ChannelSize = 10000
	}
	qlog.logChannel = make(chan LogStruct, option.ChannelSize)
	go qlog.handlerLogs()
	return qlog
}

//结束记录日志
func (qlog *QLog) Dispose() {
	qlog.run = false
}

func (qlog *QLog) handlerLogs() {
	for {
		select {
		case log := <-qlog.logChannel:
			{
				qlog.SaveLog(log.Level, log.Title, log.Data)
				break
			}
		case <-time.After(time.Second * 3):
			{
				//如果服务已经暂停
				if !qlog.run {
					//for k,v:=range ioMap{
					//	if l,ok:=v.(*log.Logger);strings.Contains(k,".log")&&ok{
					//
					//	}
					//}
					fmt.Println("log finish")
				}
			}
		}
	}

	//for log := range qlog.logChannel {
	//
	//}
}

func (qlog *QLog) Error(title string, data interface{}) {
	qlog.log("Error", title, data)
}
func (qlog *QLog) Warn(title string, data interface{}) {
	qlog.log("Warn", title, data)
}
func (qlog *QLog) Info(title string, data interface{}) {
	qlog.log("Info", title, data)
}
func (qlog *QLog) Trace(title string, data interface{}) {
	qlog.log("Trace", title, data)
}
func (qlog *QLog) Fatal(title string, data interface{}) {
	qlog.log("Fatal", title, data)
}
func (qlog *QLog) Elk(title string, data interface{}) {
	qlog.log("Elk", title, data)
}

func (qlog *QLog) log(level, title string, data interface{}) {
	qlog.logChannel <- LogStruct{
		Title: title,
		Level: level,
		Data:  data,
	}
}

//处理日志
func (qlog *QLog) SaveLog(level, title string, data interface{}) {
	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
		}
	}()
	if level == "Elk" && qlog.redisAddress != "" {
		b, _ := json.Marshal(data)
		qlog.redisClient.RPush("elk_log_queue", string(b)).Result()
		return
	}

	//写入文件
	dir := "log/" + time.Now().Format("20060102") + "/"
	if _, ok := ioMap.Load(dir); !ok {
		//如果不存在目录，则创建一条
		if _, err := os.Stat(dir); err != nil {
			if os.MkdirAll(dir, 0777) == nil {
				ioMap.Store(dir, 1)
			}
		}
	}

	fileName := dir + title + "_" + level + ".log"
	logs, ok := ioMap.Load(fileName)
	if !ok {
		file, err := os.OpenFile(fileName, os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			fmt.Println("open file error !", err)
		}
		l := log.New(file, "", log.LstdFlags)
		ioMap.Store(fileName, l)
		logs = l
	}
	if ll, ok := logs.(*log.Logger); ok && data != nil {
		ll.Println(data)
	}
}
