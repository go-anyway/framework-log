// Copyright 2025 zampo.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// @contact  zampo3380@gmail.com

package log

import (
	"context"
	"os"
	"strings"
	"sync"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	std *zap.Logger
	mu  sync.Mutex
)

// init 初始化默认的日志记录器.
func init() {
	// 初始化时使用默认配置
	std = New(NewOptions())
}

// New 根据给定的选项创建一个新的日志记录器.
func New(opts *Options) *zap.Logger {
	// 开发模式自动调整配置
	if opts.Development {
		if opts.Level == "info" {
			opts.Level = "debug"
		}
		// 开发模式下使用更友好的控制台格式
		_ = opts.Format // 避免空分支警告
		if opts.DisableCaller {
			opts.DisableCaller = false
		}
	}

	var level zapcore.Level
	// 解析日志级别字符串
	if err := level.UnmarshalText([]byte(opts.Level)); err != nil {
		// 如果解析失败，默认为 Info 级别
		level = zapcore.InfoLevel
	}

	// 配置 zap Encoder
	encoderConfig := zapcore.EncoderConfig{
		MessageKey:     "msg",
		LevelKey:       "level",
		TimeKey:        "ts",
		CallerKey:      "caller",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,    // 大写的日志级别 (INFO, ERROR)
		EncodeTime:     zapcore.ISO8601TimeEncoder,     // ISO8601 格式的时间
		EncodeDuration: zapcore.SecondsDurationEncoder, // 持续时间以秒为单位
		EncodeCaller:   zapcore.ShortCallerEncoder,     // 短格式的调用者路径 (package/file.go:line)
	}

	var encoder zapcore.Encoder
	if opts.Format == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// 创建 WriteSyncer
	ws := getWriteSyncer(opts)
	// 创建错误输出 WriteSyncer
	errorWS := getErrorWriteSyncer(opts)

	// 创建 Core
	core := zapcore.NewCore(encoder, ws, level)

	// 构建 zap 选项
	zapOpts := []zap.Option{
		zap.ErrorOutput(errorWS),
	}

	// 根据选项添加额外的 zap 选项
	if !opts.DisableCaller {
		zapOpts = append(zapOpts, zap.AddCaller())
	}
	if !opts.DisableStacktrace {
		// 开发模式下，Warn 级别及以上记录堆栈；生产模式下，Error 级别及以上记录堆栈
		stackLevel := zapcore.ErrorLevel
		if opts.Development {
			stackLevel = zapcore.WarnLevel
		}
		zapOpts = append(zapOpts, zap.AddStacktrace(stackLevel))
	}

	// 开发模式下添加开发选项
	if opts.Development {
		zapOpts = append(zapOpts, zap.Development())
	}

	// 创建 Logger
	logger := zap.New(core, zapOpts...)

	return logger
}

// getWriteSyncer 根据配置创建 zapcore.WriteSyncer.
func getWriteSyncer(opts *Options) zapcore.WriteSyncer {
	var writers []zapcore.WriteSyncer

	// 如果配置了文件名，则添加文件写入器 (使用 lumberjack 进行日志轮转)
	if opts.Filename != "" {
		lumberJackLogger := &lumberjack.Logger{
			Filename:   opts.Filename,
			MaxSize:    opts.MaxSize,
			MaxBackups: opts.MaxBackups,
			MaxAge:     opts.MaxAge,
			Compress:   opts.Compress,
		}
		writers = append(writers, zapcore.AddSync(lumberJackLogger))
	}

	// 处理控制台输出
	// 使用 map 来避免重复添加 stdout 或 stderr
	consoleWriters := make(map[string]bool)
	for _, path := range opts.OutputPaths {
		lowerPath := strings.ToLower(path)
		if _, exists := consoleWriters[lowerPath]; !exists {
			switch lowerPath {
			case "stdout":
				writers = append(writers, zapcore.AddSync(os.Stdout))
				consoleWriters[lowerPath] = true
			case "stderr":
				writers = append(writers, zapcore.AddSync(os.Stderr))
				consoleWriters[lowerPath] = true
			}
		}
	}

	return zapcore.NewMultiWriteSyncer(writers...)
}

// getErrorWriteSyncer 根据配置创建错误日志的 zapcore.WriteSyncer.
func getErrorWriteSyncer(opts *Options) zapcore.WriteSyncer {
	var writers []zapcore.WriteSyncer

	// 处理错误输出路径
	// 使用 map 来避免重复添加 stdout 或 stderr
	consoleWriters := make(map[string]bool)
	for _, path := range opts.ErrorOutputPaths {
		lowerPath := strings.ToLower(path)
		if _, exists := consoleWriters[lowerPath]; !exists {
			switch lowerPath {
			case "stdout":
				writers = append(writers, zapcore.AddSync(os.Stdout))
				consoleWriters[lowerPath] = true
			case "stderr":
				writers = append(writers, zapcore.AddSync(os.Stderr))
				consoleWriters[lowerPath] = true
			}
		}
	}

	// 如果没有配置错误输出路径，默认使用 stderr
	if len(writers) == 0 {
		writers = append(writers, zapcore.AddSync(os.Stderr))
	}

	return zapcore.NewMultiWriteSyncer(writers...)
}

// Init 使用给定的选项初始化或重新初始化全局日志记录器.
// 这个函数是线程安全的.
func Init(opts ...Option) {
	mu.Lock()
	defer mu.Unlock()
	o := NewOptions()
	o.Apply(opts...)
	std = New(o)
}

// Debug 记录一条 debug 级别的日志.
func Debug(msg string, fields ...zap.Field) {
	std.Debug(msg, fields...)
}

// Info 记录一条 info 级别的日志.
func Info(msg string, fields ...zap.Field) {
	std.Info(msg, fields...)
}

// Warn 记录一条 warn 级别的日志.
func Warn(msg string, fields ...zap.Field) {
	std.Warn(msg, fields...)
}

// Error 记录一条 error 级别的日志.
func Error(msg string, fields ...zap.Field) {
	std.Error(msg, fields...)
}

// DPanic 记录一条 dpanic 级别的日志. 在开发模式下会 panic.
func DPanic(msg string, fields ...zap.Field) {
	std.DPanic(msg, fields...)
}

// Panic 记录一条 panic 级别的日志，然后调用 panic().
func Panic(msg string, fields ...zap.Field) {
	std.Panic(msg, fields...)
}

// Fatal 记录一条 fatal 级别的日志，然后调用 os.Exit(1).
func Fatal(msg string, fields ...zap.Field) {
	std.Fatal(msg, fields...)
}

// Sync 将所有缓冲的日志条目刷新到磁盘.
// 应用程序在退出前应调用此方法.
func Sync() error {
	return std.Sync()
}

// GetLogger 返回当前的全局日志记录器.
// 这在需要传递 logger 实例而不是使用全局函数时很有用.
func GetLogger() *zap.Logger {
	return std
}

type contextKey string

const (
	traceIDKey   = contextKey("traceID")
	requestIDKey = contextKey("requestID")
)

// ContextWithTraceID 返回一个包含 traceID 的新 context.
func ContextWithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDKey, traceID)
}

// ContextWithRequestID 返回一个包含 requestID 的新 context.
func ContextWithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

// FromContext 从 context 中提取 traceID 和 requestID，返回一个包含这些字段的 Logger 实例。
// 如果上下文中没有这些值，它会返回全局的 logger。
// traceID 优先从 OpenTelemetry span 中提取，如果没有则从自定义 context key 中提取。
func FromContext(ctx context.Context) *zap.Logger {
	if ctx == nil {
		return std
	}

	var fields []zap.Field

	// 提取 traceID（优先从 OpenTelemetry span 中获取）
	traceID := extractTraceID(ctx)
	if traceID != "" {
		fields = append(fields, zap.String("traceID", traceID))
	}

	// 提取 requestID
	if requestID, ok := ctx.Value(requestIDKey).(string); ok && requestID != "" {
		fields = append(fields, zap.String("requestID", requestID))
	}

	// 如果没有字段，直接返回全局 logger，避免不必要的 With 调用
	if len(fields) == 0 {
		return std
	}

	return std.With(fields...)
}

// extractTraceID 从 context 中提取 traceID
// 优先从 OpenTelemetry span 中获取，如果没有则从自定义 context key 中获取
func extractTraceID(ctx context.Context) string {
	// 1. 尝试从 OpenTelemetry span 中获取
	if span := trace.SpanFromContext(ctx); span.SpanContext().IsValid() {
		return span.SpanContext().TraceID().String()
	}

	// 2. 尝试从自定义 context key 中获取
	if traceID, ok := ctx.Value(traceIDKey).(string); ok && traceID != "" {
		return traceID
	}

	return ""
}

// TraceIDFromContext 从 context 中提取 traceID
// 优先从 OpenTelemetry span 中获取，如果没有则从自定义 context key 中获取
func TraceIDFromContext(ctx context.Context) string {
	return extractTraceID(ctx)
}

// RequestIDFromContext 从 context 中提取 requestID
func RequestIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if requestID, ok := ctx.Value(requestIDKey).(string); ok {
		return requestID
	}
	return ""
}
