package utils

import (
	"log"
	"os"
)

// 日志参数初始化
func InitLogger(filepath string) error{
	logFile, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	log.SetOutput(logFile)
	log.SetFlags(log.Lshortfile | log.Lmicroseconds | log.Ldate)
	return nil
}
