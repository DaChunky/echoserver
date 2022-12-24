package logging

/* ----------------------------------------------------------------------------*
 * Purpose of the package is to provide a simple logging interface, which is   *
 * working out of the box.                                                     *
 * Current Issue: If you want to log to "/var/log/[application name]", the     *
 * logging directory has to be created prior because the app retrieve a        *
 * permission denied exception. Steps to create the directory:                 *
 * # cd /var/log                                                               *
 * # sudo mkdir [app name]                                                     *
 * # sudo chown $(id -g):$id -u) ./[app name]                                  *
 * # sudo chmod 744 ./[app name]                                               *
 * ----------------------------------------------------------------------------*/

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const (
	LOG_MAIN  = iota
	LOG_FATAL = iota
	LOG_ERROR = iota
	LOG_WARN  = iota
	LOG_INFO  = iota
	LOG_DEBUG = iota
)

var LogToStdOutInCaseOfError bool = false

func logLevelToString(logLevel int) string {
	ret := ""
	switch logLevel {
	case LOG_MAIN:
		ret = "MAIN"
	case LOG_FATAL:
		ret = "FATAL"
	case LOG_ERROR:
		ret = "ERROR"
	case LOG_WARN:
		ret = "WARNING"
	case LOG_INFO:
		ret = "INFO"
	case LOG_DEBUG:
		ret = "DEBUG"
	}
	return ret
}

// Logger represents a possibiliy to log messages
type Logger interface {
	Log(verbose int, msg string)
	LogFmt(verbose int, msg string, v ...interface{})
}

type logger struct {
	formatString string
	toStdOut     bool
}

func (lg *logger) Log(verbose int, msg string) {
	levelStr := logLevelToString(verbose)
	if lg.toStdOut {
		fmt.Printf("%-12s %s\n", levelStr, msg)
	} else {
		log.Printf("%-12s %s\n", levelStr, msg)
	}
}

func (lg *logger) LogFmt(verbose int, msg string, v ...interface{}) {
	lg.Log(verbose, fmt.Sprintf(msg, v...))
}

func CreateAndInitLog(file string, force bool) (Logger, error) {
	ret := new(logger)
	ret.formatString = "%s: %s"
	ret.toStdOut = false
	logPath := filepath.Dir(file)
	_, err := os.Stat(logPath)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(logPath, 0766)
			if err != nil {
				return nil, fmt.Errorf("failed to create log path '%s': %v", logPath, err)
			}
		}
	}
	lgFl, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0777)
	if err != nil {
		ret.toStdOut = LogToStdOutInCaseOfError && force
		if ret.toStdOut {
			return ret, nil
		}
		fmt.Printf("log init failed: %v\n", err)
		return nil, err
	}
	log.SetOutput(lgFl)
	return ret, nil
}

var mainLog Logger = nil

func CreateAndInitMainLog() error {
	mainAppCallSplit := strings.Split(os.Args[0], string(os.PathSeparator))
	logName := mainAppCallSplit[len(mainAppCallSplit)-1]
	logFile := fmt.Sprintf("/var/log/%s/%s.log", logName, logName)
	var err error = nil
	mainLog, err = CreateAndInitLog(logFile, false)
	if err != nil {
		fmt.Printf("create main log in var failed: %v\n", err)
		logFile = fmt.Sprintf("./%s.log", logName)
		mainLog, err = CreateAndInitLog(logFile, true)
		if err != nil {
			fmt.Printf("main log init failed: %v\n", err)
			return err
		}
	}
	mainLog.Log(LOG_INFO, "---------------------------------------")
	mainLog.Log(LOG_INFO, "application started")
	return nil
}

func checkMainLogInitialized() bool {
	if mainLog == nil {
		err := CreateAndInitMainLog()
		if err != nil {
			return false
		}
	}
	return true
}

func Log(verbose int, msg string) {
	if !checkMainLogInitialized() {
		return
	}
	mainLog.Log(verbose, msg)
}

func LogFmt(verbose int, msg string, v ...interface{}) {
	Log(verbose, fmt.Sprintf(msg, v...))
}
