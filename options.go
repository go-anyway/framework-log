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
	"fmt"

	"go.uber.org/zap/zapcore"
)

// Options 定义了日志记录器的配置项
type Options struct {
	// OutputPaths 是一个输出路径的列表，可以是 stdout, stderr, 或者文件路径
	// 默认为 ["stdout"].
	OutputPaths []string
	// ErrorOutputPaths 是一个错误日志输出路径的列表，用于记录 logger 内部的错误
	// 默认为 ["stderr"].
	ErrorOutputPaths []string
	// Level 是最低的日志记录级别.
	// 可选值: "debug", "info", "warn", "error", "dpanic", "panic", "fatal"
	// 默认为 "info".
	Level string
	// Format 指定日志的输出格式.
	// 可选值: "json", "console". 默认为 "console".
	Format string
	// DisableCaller 禁止在日志中记录调用者的文件名和行号.
	// 默认为 false.
	DisableCaller bool
	// DisableStacktrace 禁止自动捕获堆栈跟踪.
	// 默认情况下，在开发环境中，WarnLevel 及更高级别的日志会捕获堆栈，
	// 在生产环境中，ErrorLevel 及更高级别的日志会捕获堆栈.
	DisableStacktrace bool
	// Filename 是要写入日志的文件名，用于日志轮转.
	Filename string
	// MaxSize 是日志文件在轮转之前的最大大小（以MB为单位）.
	MaxSize int
	// MaxAge 是根据文件名中的时间戳保留旧日志文件的最大天数.
	MaxAge int
	// MaxBackups 是要保留的旧日志文件的最大数量.
	MaxBackups int
	// Compress 决定是否压缩轮转后的日志文件.
	// 默认为 false.
	Compress bool
	// Development 是否为开发模式.
	// 开发模式下会自动启用更详细的日志输出和堆栈跟踪.
	// 默认为 false.
	Development bool
}

// Option 是一个将配置项应用于 Options 的函数.
type Option func(*Options)

// NewOptions 创建一个带有默认值的新 Options 对象.
func NewOptions() *Options {
	return &Options{
		Level:            "info",
		Format:           "console",
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}
}

// Apply 将给定的选项应用于此 Options.
func (o *Options) Apply(opts ...Option) {
	for _, opt := range opts {
		opt(o)
	}
}

// WithLevel 设置日志级别.
// 如果提供的级别无效，将使用默认的 "info" 级别.
func WithLevel(level string) Option {
	return func(o *Options) {
		// 验证级别是否有效
		var l zapcore.Level
		if err := l.UnmarshalText([]byte(level)); err == nil {
			o.Level = level
		}
		// 如果无效，保持默认值不变
	}
}

// WithFormat 设置日志格式.
func WithFormat(format string) Option {
	return func(o *Options) {
		o.Format = format
	}
}

// WithOutputPaths 设置日志输出路径.
func WithOutputPaths(paths []string) Option {
	return func(o *Options) {
		o.OutputPaths = paths
	}
}

// WithErrorOutputPaths 设置错误日志输出路径.
func WithErrorOutputPaths(paths []string) Option {
	return func(o *Options) {
		o.ErrorOutputPaths = paths
	}
}

// WithDisableCaller 禁止调用者信息记录.
func WithDisableCaller(disable bool) Option {
	return func(o *Options) {
		o.DisableCaller = disable
	}
}

// WithDisableStacktrace 禁止堆栈跟踪.
func WithDisableStacktrace(disable bool) Option {
	return func(o *Options) {
		o.DisableStacktrace = disable
	}
}

// WithFilename 设置用于轮转的日志文件名.
func WithFilename(filename string) Option {
	return func(o *Options) {
		o.Filename = filename
	}
}

// WithMaxSize 设置轮转日志的最大大小.
func WithMaxSize(maxSize int) Option {
	return func(o *Options) {
		o.MaxSize = maxSize
	}
}

// WithMaxAge 设置轮转日志的最大保留天数.
func WithMaxAge(maxAge int) Option {
	return func(o *Options) {
		o.MaxAge = maxAge
	}
}

// WithMaxBackups 设置轮转日志的最大备份数.
func WithMaxBackups(maxBackups int) Option {
	return func(o *Options) {
		o.MaxBackups = maxBackups
	}
}

// WithCompress 设置是否压缩轮转后的日志文件.
func WithCompress(compress bool) Option {
	return func(o *Options) {
		o.Compress = compress
	}
}

// WithDevelopment 设置是否为开发模式.
func WithDevelopment(development bool) Option {
	return func(o *Options) {
		o.Development = development
	}
}

// LogConfigProvider 日志配置提供者接口
// 用于统一不同包的 LogConfig 类型转换为 log.Options
type LogConfigProvider interface {
	GetLevel() string
	GetFormat() string
	GetOutputPaths() []string
	GetErrorOutputPaths() []string
	GetDisableCaller() bool
	GetDisableStacktrace() bool
	GetFilename() string
	GetMaxSize() int
	GetMaxAge() int
	GetMaxBackups() int
	GetCompress() bool
	GetDevelopment() bool
}

// ToLogOptionsFromConfig 从 LogConfigProvider 转换为 log.Options
// 这是一个通用函数，用于统一处理不同包的 LogConfig 类型
func ToLogOptionsFromConfig(cfg LogConfigProvider) *Options {
	if cfg == nil {
		return NewOptions()
	}

	opts := NewOptions()
	opts.Level = cfg.GetLevel()
	opts.Format = cfg.GetFormat()
	opts.OutputPaths = cfg.GetOutputPaths()
	opts.ErrorOutputPaths = cfg.GetErrorOutputPaths()
	opts.DisableCaller = cfg.GetDisableCaller()
	opts.DisableStacktrace = cfg.GetDisableStacktrace()
	opts.Filename = cfg.GetFilename()
	opts.MaxSize = cfg.GetMaxSize()
	opts.MaxAge = cfg.GetMaxAge()
	opts.MaxBackups = cfg.GetMaxBackups()
	opts.Compress = cfg.GetCompress()
	opts.Development = cfg.GetDevelopment()
	return opts
}

// Config 日志配置结构体
// 用于从配置文件加载日志配置
type Config struct {
	Level             string   `yaml:"level" env:"LOG_LEVEL" default:"info"`
	Format            string   `yaml:"format" env:"LOG_FORMAT" default:"console"`
	OutputPaths       []string `yaml:"output_paths" env:"LOG_OUTPUT_PATHS"`
	ErrorOutputPaths  []string `yaml:"error_output_paths" env:"LOG_ERROR_OUTPUT_PATHS"`
	DisableCaller     bool     `yaml:"disable_caller" env:"LOG_DISABLE_CALLER" default:"false"`
	DisableStacktrace bool     `yaml:"disable_stacktrace" env:"LOG_DISABLE_STACKTRACE" default:"false"`
	Filename          string   `yaml:"filename" env:"LOG_FILENAME"`
	MaxSize           int      `yaml:"max_size" env:"LOG_MAX_SIZE" default:"100"`
	MaxAge            int      `yaml:"max_age" env:"LOG_MAX_AGE" default:"7"`
	MaxBackups        int      `yaml:"max_backups" env:"LOG_MAX_BACKUPS" default:"3"`
	Compress          bool     `yaml:"compress" env:"LOG_COMPRESS" default:"false"`
	Development       bool     `yaml:"development" env:"LOG_DEVELOPMENT" default:"false"`
}

// Validate 验证日志配置
func (c *Config) Validate() error {
	if c == nil {
		return fmt.Errorf("log config cannot be nil")
	}

	// 验证日志级别
	validLevels := map[string]bool{
		"debug": true, "info": true, "warn": true,
		"error": true, "dpanic": true, "panic": true, "fatal": true,
	}
	if c.Level != "" && !validLevels[c.Level] {
		return fmt.Errorf("log.level must be one of: debug, info, warn, error, dpanic, panic, fatal, got %s", c.Level)
	}

	// 验证日志格式
	validFormats := map[string]bool{
		"json": true, "console": true,
	}
	if c.Format != "" && !validFormats[c.Format] {
		return fmt.Errorf("log.format must be one of: json, console, got %s", c.Format)
	}

	// 验证 MaxSize
	if c.MaxSize < 0 {
		return fmt.Errorf("log.max_size must be non-negative, got %d", c.MaxSize)
	}

	// 验证 MaxAge
	if c.MaxAge < 0 {
		return fmt.Errorf("log.max_age must be non-negative, got %d", c.MaxAge)
	}

	// 验证 MaxBackups
	if c.MaxBackups < 0 {
		return fmt.Errorf("log.max_backups must be non-negative, got %d", c.MaxBackups)
	}

	return nil
}

// ToOptions 将 Config 转换为 log.Options
func (c *Config) ToOptions() *Options {
	if c == nil {
		return NewOptions()
	}
	return ToLogOptionsFromConfig(c)
}

// 实现 LogConfigProvider 接口
func (c *Config) GetLevel() string              { return c.Level }
func (c *Config) GetFormat() string             { return c.Format }
func (c *Config) GetOutputPaths() []string      { return c.OutputPaths }
func (c *Config) GetErrorOutputPaths() []string { return c.ErrorOutputPaths }
func (c *Config) GetDisableCaller() bool        { return c.DisableCaller }
func (c *Config) GetDisableStacktrace() bool    { return c.DisableStacktrace }
func (c *Config) GetFilename() string           { return c.Filename }
func (c *Config) GetMaxSize() int               { return c.MaxSize }
func (c *Config) GetMaxAge() int                { return c.MaxAge }
func (c *Config) GetMaxBackups() int            { return c.MaxBackups }
func (c *Config) GetCompress() bool             { return c.Compress }
func (c *Config) GetDevelopment() bool          { return c.Development }
