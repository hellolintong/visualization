package fileparser

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"sync"
)

var Log *zap.Logger
var once sync.Once

func init() {
	once.Do(func() {

		hook := lumberjack.Logger{
			Filename:   "visualization_log.txt", // 日志文件路径
			MaxSize:    20,                      // megabytes
			MaxBackups: 3,                       // 最多保留3个备份
			MaxAge:     7,                       //days
			Compress:   true,                    // 是否压缩 disabled by default
		}
		encoderConfig := zapcore.EncoderConfig{
			TimeKey:        "time",
			LevelKey:       "level",
			NameKey:        "internal",
			CallerKey:      "linenum",
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,  // 小写编码器
			EncodeTime:     zapcore.ISO8601TimeEncoder,     // ISO8601 UTC 时间格式
			EncodeDuration: zapcore.SecondsDurationEncoder, //
			EncodeCaller:   zapcore.FullCallerEncoder,      // 全路径编码器
			EncodeName:     zapcore.FullNameEncoder,
		}

		// 设置日志级别
		atomicLevel := zap.NewAtomicLevel()
		atomicLevel.SetLevel(zap.InfoLevel)

		core := zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),                                           // 编码器配置
			zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(&hook)), // 打印到控制台和文件
			atomicLevel,                                                                     // 日志级别
		)

		// 开启开发模式，堆栈跟踪
		caller := zap.AddCaller()
		// 开启文件及行号
		development := zap.Development()
		// 设置初始化字段
		filed := zap.Fields()
		// 构造日志
		Log = zap.New(core, caller, development, filed)
	})
}
