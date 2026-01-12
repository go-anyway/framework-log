package log_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/go-anyway/framework-log"

	"go.uber.org/zap"
)

// TestDefaultLogging 测试默认日志记录器的输出.
// 这个测试主要作为基本的功能验证和用法示例.
func TestDefaultLogging(t *testing.T) {
	// log 包中的 init() 函数会初始化一个默认的记录器，所以我们可以直接使用.
	log.Info("这是来自默认记录器的一条 info 消息。")
	log.Warn("这是一条 warning 消息。", zap.String("key", "value"))
	log.Error("这是一条 error 消息。", zap.Int("code", 123))

	// 如果没有发生 panic，测试即通过.
	// 在测试中捕获 stdout/stderr 相对复杂，因此这里主要作为冒烟测试.
}

// TestInitWithLevel 测试使用特定级别重新初始化记录器.
func TestInitWithLevel(t *testing.T) {
	// 使用更高的日志级别 "warn" 进行初始化.
	// 这样，"debug" 和 "info" 级别的日志将不会被打印.
	log.Init(log.WithLevel("warn"))

	log.Debug("这条 debug 消息不应该显示。")
	log.Info("这条 info 消息不应该显示。")
	log.Warn("这条 warning 消息应该会显示。")
	log.Error("这条 error 消息应该会显示。")

	// 为不影响其他测试，恢复默认日志级别.
	defer log.Init(log.WithLevel("info"))
}

// TestFileLogging 测试日志记录到文件.
func TestFileLogging(t *testing.T) {
	logFile := "test.log"
	// 测试结束后清理日志文件.
	defer os.Remove(logFile)

	// 初始化记录器，使其写入到文件，并使用 JSON 格式.
	log.Init(
		log.WithLevel("debug"),
		log.WithFormat("json"),
		log.WithFilename(logFile),
		// 将 OutputPaths 设置为文件，而不是默认的 "stdout"
		// 注意：如果想同时输出到文件和控制台，可以设置为 []string{logFile, "stdout"}
		log.WithOutputPaths([]string{}),
	)

	log.Info("这条消息应该被写入测试日志文件。", zap.String("test_name", "TestFileLogging"))
	log.Error("这是一个文件错误日志。", zap.Time("time", time.Now()))

	// 在继续之前，确保所有缓冲的日志都已写入磁盘.
	err := log.Sync()
	if err != nil {
		t.Errorf("同步 logger 失败: %v", err)
	}

	// 检查日志文件是否已创建且不为空.
	info, err := os.Stat(logFile)
	if os.IsNotExist(err) {
		t.Fatalf("日志文件 '%s' 未被创建。", logFile)
	}
	if info.Size() == 0 {
		t.Fatalf("日志文件 '%s' 为空。", logFile)
	}
	t.Logf("日志文件 '%s' 已成功创建，大小为 %d 字节。", logFile, info.Size())
}

// TestFromContext 测试从 context 创建日志记录器.
func TestFromContext(t *testing.T) {
	// 创建一个父 context
	ctx := context.Background()

	// 定义 traceID 和 requestID
	traceID := "my-trace-id-123"
	requestID := "my-request-id-456"

	// 使用辅助函数将 ID 添加到 context 中
	ctx = log.ContextWithTraceID(ctx, traceID)
	ctx = log.ContextWithRequestID(ctx, requestID)

	// 从 context 获取日志记录器
	logger := log.FromContext(ctx)

	// 使用这个 logger 记录的日志会自动包含 traceID 和 requestID 字段
	// 注意：在标准测试输出中，这些字段会以 console 格式显示
	logger.Info("这是一条来自上下文日志记录器的 info 消息。")
	logger.Warn("这是来自相同上下文的另一条 warn 消息。")

	// 测试一个没有上下文的 logger
	loggerNoContext := log.FromContext(context.Background())
	loggerNoContext.Info("这条日志不应该包含 traceID 或 requestID。")

	// 为了不影响其他测试，恢复默认日志级别.
	log.Init(log.WithLevel("info"))
}

// TestOptions_DefaultValues 测试 Options 的默认值.
func TestOptions_DefaultValues(t *testing.T) {
	opts := log.NewOptions()

	if opts.Level != "info" {
		t.Errorf("Default Level = %s, want info", opts.Level)
	}
	if opts.Format != "console" {
		t.Errorf("Default Format = %s, want console", opts.Format)
	}
	if len(opts.OutputPaths) != 1 || opts.OutputPaths[0] != "stdout" {
		t.Errorf("Default OutputPaths = %v, want [stdout]", opts.OutputPaths)
	}
	if len(opts.ErrorOutputPaths) != 1 || opts.ErrorOutputPaths[0] != "stderr" {
		t.Errorf("Default ErrorOutputPaths = %v, want [stderr]", opts.ErrorOutputPaths)
	}
	if opts.DisableCaller {
		t.Errorf("Default DisableCaller = %v, want false", opts.DisableCaller)
	}
	if opts.DisableStacktrace {
		t.Errorf("Default DisableStacktrace = %v, want false", opts.DisableStacktrace)
	}
	if opts.Filename != "" {
		t.Errorf("Default Filename = %s, want empty", opts.Filename)
	}
	if opts.MaxSize != 0 {
		t.Errorf("Default MaxSize = %d, want 0", opts.MaxSize)
	}
	if opts.MaxAge != 0 {
		t.Errorf("Default MaxAge = %d, want 0", opts.MaxAge)
	}
	if opts.MaxBackups != 0 {
		t.Errorf("Default MaxBackups = %d, want 0", opts.MaxBackups)
	}
	if opts.Compress {
		t.Errorf("Default Compress = %v, want false", opts.Compress)
	}
	if opts.Development {
		t.Errorf("Default Development = %v, want false", opts.Development)
	}
}

// TestWithOptions 测试各种 Option 函数.
func TestWithOptions(t *testing.T) {
	opts := log.NewOptions()

	log.WithLevel("debug")(opts)
	if opts.Level != "debug" {
		t.Errorf("WithLevel() Level = %s, want debug", opts.Level)
	}

	log.WithFormat("json")(opts)
	if opts.Format != "json" {
		t.Errorf("WithFormat() Format = %s, want json", opts.Format)
	}

	log.WithOutputPaths([]string{"stdout", "file.log"})(opts)
	if len(opts.OutputPaths) != 2 {
		t.Errorf("WithOutputPaths() OutputPaths length = %d, want 2", len(opts.OutputPaths))
	}

	log.WithErrorOutputPaths([]string{"stderr"})(opts)
	if len(opts.ErrorOutputPaths) != 1 {
		t.Errorf("WithErrorOutputPaths() ErrorOutputPaths length = %d, want 1", len(opts.ErrorOutputPaths))
	}

	log.WithDisableCaller(true)(opts)
	if !opts.DisableCaller {
		t.Errorf("WithDisableCaller() DisableCaller = %v, want true", opts.DisableCaller)
	}

	log.WithDisableStacktrace(true)(opts)
	if !opts.DisableStacktrace {
		t.Errorf("WithDisableStacktrace() DisableStacktrace = %v, want true", opts.DisableStacktrace)
	}

	log.WithFilename("app.log")(opts)
	if opts.Filename != "app.log" {
		t.Errorf("WithFilename() Filename = %s, want app.log", opts.Filename)
	}

	log.WithMaxSize(100)(opts)
	if opts.MaxSize != 100 {
		t.Errorf("WithMaxSize() MaxSize = %d, want 100", opts.MaxSize)
	}

	log.WithMaxAge(7)(opts)
	if opts.MaxAge != 7 {
		t.Errorf("WithMaxAge() MaxAge = %d, want 7", opts.MaxAge)
	}

	log.WithMaxBackups(5)(opts)
	if opts.MaxBackups != 5 {
		t.Errorf("WithMaxBackups() MaxBackups = %d, want 5", opts.MaxBackups)
	}

	log.WithCompress(true)(opts)
	if !opts.Compress {
		t.Errorf("WithCompress() Compress = %v, want true", opts.Compress)
	}

	log.WithDevelopment(true)(opts)
	if !opts.Development {
		t.Errorf("WithDevelopment() Development = %v, want true", opts.Development)
	}
}

// TestOptions_Apply 测试 Options.Apply 方法.
func TestOptions_Apply(t *testing.T) {
	opts := log.NewOptions()
	opts.Apply(
		log.WithLevel("error"),
		log.WithFormat("json"),
		log.WithFilename("error.log"),
	)

	if opts.Level != "error" {
		t.Errorf("Apply() Level = %s, want error", opts.Level)
	}
	if opts.Format != "json" {
		t.Errorf("Apply() Format = %s, want json", opts.Format)
	}
	if opts.Filename != "error.log" {
		t.Errorf("Apply() Filename = %s, want error.log", opts.Filename)
	}
}

// TestWithInvalidLevel 测试无效日志级别时的行为.
func TestWithInvalidLevel(t *testing.T) {
	opts := log.NewOptions()
	originalLevel := opts.Level

	log.WithLevel("invalid-level")(opts)

	if opts.Level != originalLevel {
		t.Errorf("WithInvalidLevel() Level changed from %s, want stay %s", originalLevel, originalLevel)
	}
}

// TestDevelopmentMode 测试开发模式配置.
func TestDevelopmentMode(t *testing.T) {
	log.Init(
		log.WithDevelopment(true),
		log.WithLevel("info"),
	)

	logger := log.GetLogger()
	if logger == nil {
		t.Error("GetLogger() returned nil")
	}

	log.Init(log.WithLevel("info"))
}

// TestConsoleFormat 测试控制台格式输出.
func TestConsoleFormat(t *testing.T) {
	log.Init(
		log.WithFormat("console"),
		log.WithLevel("debug"),
	)

	log.Debug("Debug message in console format")
	log.Info("Info message in console format")

	log.Init(log.WithLevel("info"))
}

// TestGetLogger 测试获取全局日志记录器.
func TestGetLogger(t *testing.T) {
	logger := log.GetLogger()

	if logger == nil {
		t.Error("GetLogger() returned nil")
	}

	if logger == zap.NewNop() {
		t.Error("GetLogger() should not return nop logger")
	}
}

// TestContextWithNilContext 测试 nil context 的处理.
func TestContextWithNilContext(t *testing.T) {
	logger := log.FromContext(nil)

	if logger == nil {
		t.Error("FromContext(nil) returned nil")
	}
}

// TestTraceIDFromContext 测试从 context 获取 traceID.
func TestTraceIDFromContext(t *testing.T) {
	ctx := context.Background()

	traceID := log.TraceIDFromContext(ctx)
	if traceID != "" {
		t.Errorf("TraceIDFromContext(empty context) = %s, want empty", traceID)
	}

	ctx = log.ContextWithTraceID(ctx, "test-trace-id")
	traceID = log.TraceIDFromContext(ctx)
	if traceID != "test-trace-id" {
		t.Errorf("TraceIDFromContext() = %s, want test-trace-id", traceID)
	}
}

// TestRequestIDFromContext 测试从 context 获取 requestID.
func TestRequestIDFromContext(t *testing.T) {
	ctx := context.Background()

	requestID := log.RequestIDFromContext(ctx)
	if requestID != "" {
		t.Errorf("RequestIDFromContext(empty context) = %s, want empty", requestID)
	}

	ctx = log.ContextWithRequestID(ctx, "test-request-id")
	requestID = log.RequestIDFromContext(ctx)
	if requestID != "test-request-id" {
		t.Errorf("RequestIDFromContext() = %s, want test-request-id", requestID)
	}
}

// TestContextWithEmptyTraceID 测试空 traceID 的处理.
func TestContextWithEmptyTraceID(t *testing.T) {
	ctx := context.Background()
	ctx = log.ContextWithTraceID(ctx, "")

	logger := log.FromContext(ctx)
	if logger == nil {
		t.Error("FromContext() with empty traceID returned nil")
	}
}

// TestContextWithEmptyRequestID 测试空 requestID 的处理.
func TestContextWithEmptyRequestID(t *testing.T) {
	ctx := context.Background()
	ctx = log.ContextWithRequestID(ctx, "")

	logger := log.FromContext(ctx)
	if logger == nil {
		t.Error("FromContext() with empty requestID returned nil")
	}
}

// TestLogLevels 测试所有日志级别.
func TestLogLevels(t *testing.T) {
	log.Init(log.WithLevel("debug"))

	log.Debug("Debug message")
	log.Info("Info message")
	log.Warn("Warning message")
	log.Error("Error message")

	log.Init(log.WithLevel("info"))
}

// TestSyncOnDefaultLogger 测试默认记录器的同步操作.
// 注意：在某些测试环境中，stdout 可能无法正常同步，这是已知的限制。
// 测试主要验证 Sync() 方法不会 panic。
func TestSyncOnDefaultLogger(t *testing.T) {
	err := log.Sync()
	// stdout 在某些测试环境中可能无法正常同步，这是已知的限制
	// 我们不将这种错误视为测试失败
	if err != nil {
		t.Logf("Sync() returned error (在测试环境中可能发生): %v", err)
	}
}

// TestFileRotationConfig 测试日志轮转配置.
func TestFileRotationConfig(t *testing.T) {
	logFile := "rotation_test.log"
	defer os.Remove(logFile)

	log.Init(
		log.WithFilename(logFile),
		log.WithMaxSize(1),
		log.WithMaxAge(1),
		log.WithMaxBackups(3),
		log.WithCompress(true),
		log.WithOutputPaths([]string{}),
	)

	log.Info("Test message for rotation config")

	err := log.Sync()
	if err != nil {
		t.Errorf("Sync() error: %v", err)
	}
}
