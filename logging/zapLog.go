// Copyright 2020 newtbig Author. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package logging

import (
	"fmt"
	"io"
	"os"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.SugaredLogger

func Panic(msg string) {
	time.Sleep(time.Second * 5)
	panic(msg)
}

func Debug(args ...interface{}) {
	Logger.Debug(args)
}

func Info(args ...interface{}) {
	Logger.Info(args)
}

func Warn(args ...interface{}) {
	Logger.Warn(args)
}

func Error(args ...interface{}) {
	Logger.Error(args)
}

func Fatal(args ...interface{}) {
	Logger.Fatal(args)
}

const (
	out_log_name = "out"
)

func InitLog(logPath, logName string, debug bool) {
	encoder := zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
		MessageKey:    "msg",
		LevelKey:      "level",
		TimeKey:       "ts",
		CallerKey:     "caller", //"file",
		StacktraceKey: "trace",
		LineEnding:    zapcore.DefaultLineEnding,
		EncodeLevel:   zapcore.LowercaseLevelEncoder,
		//EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeCaller: zapcore.ShortCallerEncoder,
		EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Format("2006-01-02 15:04:05"))
		},
		//EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeDuration: func(d time.Duration, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendInt64(int64(d) / 1000000)
		},
	})

	infoLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return true
	})

	warnLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.WarnLevel
	})

	if logName == "" {
		logName = out_log_name
	}
	infoHook := os.Stdout
	zc := zapcore.NewCore(encoder, zapcore.AddSync(infoHook), infoLevel)
	if !debug {
		infoLogName := fmt.Sprintf("/%s.log", logName)
		infoHook_1 := getWriter(logPath, infoLogName)
		infoLevel = zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return lvl >= zapcore.InfoLevel
		})
		zc = zapcore.NewCore(encoder, zapcore.AddSync(infoHook_1), infoLevel)
	}

	errLogName := fmt.Sprintf("/%s.err", logName)
	errorHook := getWriter(logPath, errLogName)

	core := zapcore.NewTee(
		zc,
		zapcore.NewCore(encoder, zapcore.AddSync(errorHook), warnLevel),
	)

	log := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel))
	Logger = log.Sugar()
	defer Logger.Sync()
}

func getWriter(logPath, filename string) io.Writer {
	hook, err := rotatelogs.New(
		logPath+filename+".%Y%m%d",
		rotatelogs.WithLinkName(filename),
		rotatelogs.WithMaxAge(time.Hour*24*7),
		rotatelogs.WithRotationTime(time.Hour*24),
	)
	if err != nil {
		panic(err)
	}
	return hook
}
