# 日志 Context 显式传递功能

## 功能说明

本功能要求所有需要包含 `traceID` 和 `requestID` 的日志调用，必须使用 `log.FromContext(ctx)` 显式传递 context。这种方式更直观、更符合工程化实践，不需要在 goroutine 间保存 context。

## 工作原理

1. **显式传递**: 所有日志调用必须显式传递 context，使用 `log.FromContext(ctx)`
2. **自动提取**: `FromContext` 函数会自动从 context 中提取 `traceID` 和 `requestID`
3. **OpenTelemetry 支持**: 优先从 OpenTelemetry span 中提取 traceID，如果没有则从自定义 context key 中提取

## 使用方式

### 标准用法（推荐）

**所有需要 traceID 和 requestID 的日志调用，必须使用 `log.FromContext(ctx)`**：

```go
// 方式 1: 直接使用 FromContext（推荐）
log.FromContext(ctx).Info("message", zap.String("key", "value"))

// 方式 2: 先获取 logger，再使用
logger := log.FromContext(ctx)
logger.Info("message", zap.String("key", "value"))
```

### 不需要 traceID 的日志

对于不需要 traceID 和 requestID 的日志（如启动日志、配置日志等），可以直接使用全局函数：

```go
log.Info("Application started")
log.Warn("Config not found", zap.Error(err))
```

## TraceID 提取优先级

1. **OpenTelemetry Span**: 优先从 OpenTelemetry span 中提取 traceID
2. **自定义 Context Key**: 如果 OpenTelemetry span 不存在，从自定义 context key 中提取（通过 `log.ContextWithTraceID` 设置）

## RequestID 提取

从自定义 context key 中提取（通过 `log.ContextWithRequestID` 设置）。

## 示例

### 正确用法

```go
// 在业务代码中
func (s *Service) ProcessRequest(ctx context.Context, req *Request) error {
    // 使用 FromContext 显式传递 context
    log.FromContext(ctx).Info("Processing request",
        zap.String("request_id", req.ID),
    )

    // 处理业务逻辑
    result, err := s.process(ctx, req)
    if err != nil {
        log.FromContext(ctx).Error("Failed to process request",
            zap.Error(err),
        )
        return err
    }

    log.FromContext(ctx).Info("Request processed successfully",
        zap.String("result", result),
    )
    return nil
}
```

### 消息队列消费者

```go
// Kafka 消费者
for msg := range messages {
    // 从消息头中提取 trace context
    msgCtx := kafka.ExtractTraceContextFromMessage(msg)
    // 使用 FromContext 显式传递 context
    log.FromContext(msgCtx).Info("Processing message",
        zap.String("topic", msg.Topic),
    )
    processMessage(msgCtx, msg.Value)
}

// RabbitMQ 消费者
for delivery := range deliveries {
    // 从消息头中提取 trace context
    msgCtx := mq.ExtractTraceContextFromDelivery(delivery)
    // 使用 FromContext 显式传递 context
    log.FromContext(msgCtx).Info("Processing message",
        zap.String("exchange", delivery.Exchange),
    )
    processMessage(msgCtx, delivery.Body)
}
```

## 日志输出示例

### 使用 FromContext（包含 traceID 和 requestID）

```json
{
  "level": "INFO",
  "ts": "2026-01-05T15:42:27.961+0800",
  "caller": "service/api_service.go:1234",
  "msg": "Processing request",
  "request_id": "req-123",
  "traceID": "4bf92f3577b34da6a3ce929d0e0e4736",
  "requestID": "550e8400-e29b-41d4-a716-446655440000"
}
```

### 直接使用全局函数（不包含 traceID 和 requestID）

```json
{
  "level": "INFO",
  "ts": "2026-01-05T15:42:27.961+0800",
  "caller": "cmd/main.go:45",
  "msg": "Application started"
}
```

## 优势

1. **显式传递**: 所有 context 传递都是显式的，代码更清晰
2. **无状态**: 不需要在 goroutine 间保存 context，避免内存泄漏
3. **工程化**: 符合 Go 语言的最佳实践，context 应该显式传递
4. **易于调试**: 可以清楚地看到每个日志调用使用的 context

## 注意事项

1. **必须显式传递**: 所有需要 traceID 和 requestID 的日志调用，必须使用 `log.FromContext(ctx)`
2. **Context 传递**: 确保 context 在调用链中正确传递
3. **性能**: `FromContext` 函数性能开销很小，可以放心使用
