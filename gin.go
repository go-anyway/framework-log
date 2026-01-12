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
	"io"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// GinWriter 返回一个 io.Writer，用于将 gin 的日志输出到 zap logger.
// 这个 Writer 会将所有写入的内容作为 Info 级别的日志记录.
func GinWriter() io.Writer {
	return &zapWriter{
		logger: std,
		level:  zapcore.InfoLevel,
	}
}

// GinErrorWriter 返回一个 io.Writer，用于将 gin 的错误日志输出到 zap logger.
// 这个 Writer 会将所有写入的内容作为 Error 级别的日志记录.
func GinErrorWriter() io.Writer {
	return &zapWriter{
		logger: std,
		level:  zapcore.ErrorLevel,
	}
}

// zapWriter 实现了 io.Writer 接口，用于将标准输出重定向到 zap logger.
type zapWriter struct {
	logger *zap.Logger
	level  zapcore.Level
}

// Write 实现 io.Writer 接口.
func (w *zapWriter) Write(p []byte) (n int, err error) {
	// 移除末尾的换行符（zap 会自动添加）
	msg := string(p)
	if len(msg) > 0 && msg[len(msg)-1] == '\n' {
		msg = msg[:len(msg)-1]
	}

	// 根据级别记录日志
	switch w.level {
	case zapcore.DebugLevel:
		w.logger.Debug(msg)
	case zapcore.InfoLevel:
		w.logger.Info(msg)
	case zapcore.WarnLevel:
		w.logger.Warn(msg)
	case zapcore.ErrorLevel:
		w.logger.Error(msg)
	default:
		w.logger.Info(msg)
	}

	return len(p), nil
}
