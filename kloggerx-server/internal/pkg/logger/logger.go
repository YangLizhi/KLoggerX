package logger

import (
	"kloggerx-server/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var logger *zap.SugaredLogger

func Init(cfg config.LogConfig) {
	writeSyncer := zapcore.AddSync(&lumberjack.Logger{
		Filename:   cfg.Filename,
		MaxSize:    cfg.MaxSize,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAge,
	})

	var level zapcore.Level
	switch cfg.Level {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	default:
		level = zapcore.InfoLevel
	}

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		writeSyncer,
		level,
	)
	l := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	logger = l.Sugar()
}

func Info(args ...interface{})              { if logger != nil { logger.Info(args...) } }
func Infof(tpl string, args ...interface{}) { if logger != nil { logger.Infof(tpl, args...) } }
func Error(args ...interface{})             { if logger != nil { logger.Error(args...) } }
func Errorf(tpl string, args ...interface{}){ if logger != nil { logger.Errorf(tpl, args...) } }
func Warn(args ...interface{})              { if logger != nil { logger.Warn(args...) } }
func Debug(args ...interface{})             { if logger != nil { logger.Debug(args...) } }
