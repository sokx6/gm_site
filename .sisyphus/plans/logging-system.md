# Go 后端日志系统

## TL;DR

> **Quick Summary**: 零外部依赖，基于 Go `log/slog` 构建生产级日志系统——彩色控制台 + JSON 文件双输出，自定义错误链，全量替换裸 `log.*` 和隐藏错误的 handler。

> **Deliverables**:
> - `internal/logger` — slog 初始化（彩色 ANSI 控制台 + JSON 文件）
> - `internal/apperror` — 自定义错误类型（含 Code/Msg/Cause 链）
> - Echo 中间件 — 将 `*slog.Logger` 注入请求上下文
> - 全量替换 — 12 处 `log.*`、6 处 `c.Logger()`、30 处 handler 隐藏错误

> **Estimated Effort**: Large
> **Parallel Execution**: YES — 2 waves
> **Critical Path**: Task 1-2 → Task 3 → Task 4-7 (parallel) → Task 8

---

## Context

### Original Request
后端需要：彩色日志输出、文件持久化、自定义错误类型保证错误链完整可查。

### Interview Summary
- Go 版本 1.25.3——`slog` 完全可用，`DiscardHandler` 可用，`NewMultiHandler` 需等 1.26
- 当前 ~76 处 `fmt.Errorf`，12 处裸 `log.*`，30 处 handler 吞错
- 用户偏好零外部依赖

### Metis Review
- ReplaceAttr 用于脱敏（密码/token）
- 每个请求创建 request-scoped logger（含 requestID, method, path）
- repository 层错误包装已较好（`fmt.Errorf("ctx: %w", err)`），保持不动

---

## Work Objectives

### Core Objective
构建生产级日志系统：所有错误链可追踪，控制台彩色可读，文件持久化可查。

### Must Have
- 零外部依赖（纯 Go 标准库 + ANSI escape codes）
- 控制台彩色输出（DEBUG 青 / INFO 绿 / WARN 黄 / ERROR 红）
- 文件 JSON 格式持久化
- 自定义错误类型含 Code + Message + Cause（Unwrap 链）
- Echo 中间件注入 request-scoped logger
- 所有 handler 在返回前先 log 原始错误

### Must NOT Have
- 不引入第三方日志库（zerolog/zap）
- 不引入第三方错误库
- 不动前端代码
- 不动数据库 schema

---

## Verification Strategy

- **Automated tests**: TESTS-AFTER
- **QA**: `go build ./...` 通过，后端启动后观察控制台彩色输出，检查文件日志

---

## Execution Strategy

```
Wave 1 (Start Immediately — foundation):
├── Task 1: internal/apperror — custom error types [deep]
├── Task 2: internal/logger — slog setup + ANSI Handler [deep]
└── Task 3: Echo middleware — logger injection [quick]

Wave 2 (After Wave 1 — parallel migration):
├── Task 4: handler 层 — 30 处 Error(c) 加日志 [unspecified-high]
├── Task 5: service 层 — log.* 替换为 slog [quick]
├── Task 6: cmd/server — 启动日志用 slog [quick]
├── Task 7: websocket — log.* 替换 [quick]

Wave FINAL:
└── Task 8: go build + 启动验证
```

---

## TODOs

---

- [x] 1. `internal/apperror` — 自定义错误类型

  **What to do**:
  1. 新建 `backend/internal/apperror/apperror.go`
  2. 定义 `AppError` 结构体：
     ```go
     type AppError struct {
         Code    int    `json:"code"`    // HTTP status code
         Message string `json:"message"` // 用户友好消息
         Cause   error  `json:"-"`       // 原始错误（不序列化）
     }
     func (e *AppError) Error() string { ... }
     func (e *AppError) Unwrap() error { return e.Cause }
     ```
  3. 工厂函数：`NewNotFound(msg string, cause error)`、`NewValidation(msg string, cause error)`、`NewAuth(msg string, cause error)`、`NewInternal(msg string, cause error)`
  4. 辅助：`Code(err error) int` — 从 error chain 中提取 HTTP 状态码（默认 500）
  5. 新建 `backend/internal/apperror/apperror_test.go` — 验证 Unwrap 链

  **Must NOT do**: 不引入外部依赖

  **QA Scenarios**: `go test ./internal/apperror/...` — 验证 error wrapping 链、Code 提取

  **Commit**: `feat(backend): add apperror package with custom error types`

- [x] 2. `internal/logger` — slog 初始化 + ANSI Handler

  **What to do**:
  1. 新建 `backend/internal/logger/logger.go`
  2. 定义 ANSI 颜色常量：`colorDebug="\033[36m"`, `colorInfo="\033[32m"`, `colorWarn="\033[33m"`, `colorError="\033[31m"`, `colorReset="\033[0m"`
  3. 实现 `coloredHandler` 结构体（实现 `slog.Handler` 接口：`Enabled`, `Handle`, `WithAttrs`, `WithGroup`）——Handle 方法中根据 level 选择颜色写入 `io.Writer`
  4. `Init(logFilePath string) *slog.Logger`——创建双输出：
     - 控制台：`coloredHandler{writer: os.Stdout, level: slog.LevelDebug}`
     - 文件：`slog.NewJSONHandler(fileWriter, &slog.HandlerOptions{Level: slog.LevelInfo})`
     - 用自定义 `multiHandler`（手动实现 fan-out，Go 1.25 无 NewMultiHandler）
  5. 全局变量 `var L *slog.Logger` 在 `Init` 中赋值
  6. `ReplaceAttr` 用于脱敏（密码字段替换为 `[REDACTED]`）

  **Must NOT do**: 不引入第三方日志库

  **QA Scenarios**: 启动后端，观察控制台彩色输出；tail JSON 文件验证格式正确

  **Commit**: `feat(backend): add slog-based logger with ANSI console + JSON file`

- [x] 3. Echo 中间件 — logger 注入上下文

  **What to do**:
  1. 新建 `backend/internal/middleware/logger.go`
  2. 定义中间件函数 `LoggerMiddleware(logger *slog.Logger) echo.MiddlewareFunc`
  3. 每个请求开始时：`requestLogger := logger.With("request_id", uuid, "method", c.Request().Method, "path", c.Path())`
  4. 注入：`c.Set("logger", requestLogger)`
  5. 请求结束时：`requestLogger.Info("request completed", "status", c.Response().Status, "duration", latency)`
  6. 在 `cmd/server/main.go` 中 `e.Use(middleware.LoggerMiddleware(logger.L))`
  7. 导出 `func GetLogger(c echo.Context) *slog.Logger` —— 从 context 取 logger（带 fallback 到全局 L）

  **Must NOT do**: 不改变现有中间件的顺序或功能

  **QA Scenarios**: 发 HTTP 请求，检查控制台和文件日志中有 request_id 字段

  **Commit**: `feat(backend): add Echo logger middleware`

- [x] 4. Handler 层 — 错误日志化

  **What to do**:
  1. 改造 `backend/internal/handler/image.go` — 所有 `return Error(c, ...)` 前加 `logger.Error(...)`
  2. 改造 `backend/internal/handler/auth.go` — 同上，尤其注册邮件通知 goroutine 内加日志
  3. 改造 `backend/internal/handler/comment.go` — 同上
  4. 改造 `backend/internal/handler/album.go` — 同上
  5. 改造 `backend/internal/handler/admin.go` — 同上
  6. 模式：在每个 `return Error(c, ...)` 前加一行：
     ```go
     logger := logger.GetLogger(c)
     logger.Error("operation failed", "err", err, "detail", "具体操作描述")
     ```
  7. 同时在 `response.go` 的 `Error()` 函数中加 `c.Logger()` fallback（如果 context 无 logger）

  **Must NOT do**: 不改 repository 或 service 层（那些已有较好的错误包装）

  **QA Scenarios**: go build 通过；发请求触发出错场景，检查日志中是否输出完整错误链

  **Commit**: `refactor(backend): add slog error logging to all handlers`

- [x] 5. Service 层 — 替换裸 log

  **What to do**:
  1. `internal/service/stats.go` — `log.Printf` → `logger.L.Error(...)` / `logger.L.Warn(...)`
  2. `internal/service/email.go` — `log.Printf` → `logger.L.Error(...)`（MockEmail 的日志也替换）

  **Must NOT do**: 不改变业务逻辑

  **QA Scenarios**: go build 通过

  **Commit**: `refactor(backend): migrate service layer to slog`

- [x] 6. `cmd/server/main.go` — 启动日志用 slog

  **What to do**:
  1. 在 `main()` 最开始调用 `logger.Init("/var/log/gm_site/gm_site.log")`（目录不存在则 os.MkdirAll）
  2. 现有 `log.Fatalf` → `logger.L.Error(...)` + `os.Exit(1)`
  3. 现有 `log.Println("Database connected")` → `logger.L.Info("database connected")`
  4. Echo 的启动日志替换：`e.Logger` 改用 slog wrapper
  5. Graceful shutdown 时调用 sh

  **Must NOT do**: 不改变 panic/recover 行为

  **QA Scenarios**: 启动后端，观察控制台彩色启动日志

  **Commit**: `refactor(backend): migrate cmd/server to slog`

- [x] 7. WebSocket — 替换裸 log

  **What to do**:
  1. `internal/websocket/client.go` — `log.Printf` → `logger.L.Error(...)`

  **Must NOT do**: 不改变 WebSocket 连接逻辑

  **QA Scenarios**: go build 通过

  **Commit**: `refactor(backend): migrate websocket to slog`

- [x] 8. 集成验证

  **What to do**:
  1. `go build ./...` — 零错误
  2. `go vet ./...` — 零警告
  3. 启动后端 → 确认控制台彩色输出
  4. `tail /var/log/gm_site/gm_site.log` → 确认 JSON 格式日志
  5. 触发注册 → 确认日志中含 request_id 和完整错误链

  **QA Scenarios**: 以上全部验证

  **Commit**: none

---

## Final Verification Wave

---

## Commit Strategy

- **1**: `feat(backend): add apperror package with custom error types`
- **2**: `feat(backend): add slog-based logger with ANSI console + JSON file`
- **3**: `feat(backend): add Echo logger middleware`
- **4-7**: `refactor(backend): migrate to slog — handlers/services/cmd/ws`
- **8**: `verify: go build passes, colored log output confirmed`

---

## Success Criteria
- [ ] `go build ./...` 零错误
- [ ] 控制台日志按级别分色
- [ ] 文件日志为有效 JSON
- [ ] 错误返回含完整 Unwrap 链
