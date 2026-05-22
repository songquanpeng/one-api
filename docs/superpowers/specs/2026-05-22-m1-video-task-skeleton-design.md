# M1 视频任务骨架产品方案与原型设计

日期：2026-05-22

## 背景

M0 本地 demo 和二开基线已完成。M1 的目标不是接入真实 Seedance，也不是做完整视频创作产品，而是先证明帧桥 API 能承载“异步视频任务”这一新的业务形态。

M1 必须回答的问题：

- 开发者能否通过 API 创建、查询、取消视频任务。
- 系统能否把任务从创建推进到运行、成功、失败或取消。
- 管理员能否在后台看到任务当前状态、请求参数、fake provider 信息和状态事件。
- 后续 M2 计费、M3 真实 Seedance、M4 Playground 能否在这个骨架上继续演进。

## 产品定位

M1 采用“API + 运维后台”方案。

包含：

- 面向开发者的 3 个视频任务 API。
- 管理员后台的视频任务列表页。
- 管理员后台的视频任务详情页。
- 管理员后台的调试创建任务表单。
- fake provider，支持模拟成功、失败、慢任务和取消。

不包含：

- 普通用户视频创作 Playground。
- 真实 Seedance 上游接入。
- 真实扣费、结算、退款和账本。
- 对象存储转存。
- webhook 通知。
- 多上游容灾和生产级重试。

M1 后台调试表单不是正式用户入口，只用于开发和验收，避免每次都依赖 curl。

## 用户角色

### 开发者

开发者通过 API 创建、查询和取消视频任务。M1 阶段开发者拿到的是 fake provider 结果，用来验证协议和任务生命周期。

### 管理员

管理员在后台查看所有视频任务，排查任务状态、错误信息和事件流。管理员也可以通过后台调试表单创建 fake 任务。

### 普通用户

普通用户在 M1 不获得独立视频创作页面。正式创作体验放到 M4 Playground。

## API 设计

对外 API 使用帧桥自己的稳定抽象，不直接暴露 Seedance 上游字段。M1 内部 provider 为 `fake`，M3 接真实 Seedance 时不改变开发者调用方式。

### 创建视频任务

```text
POST /v1/videos/generations
Authorization: Bearer <token>
```

请求体：

```json
{
  "model": "fake-seedance-lite",
  "prompt": "一只机械手在桌面拼装芯片",
  "input_image_url": null,
  "duration": 5,
  "aspect_ratio": "16:9",
  "resolution": "720p",
  "test_mode": "success"
}
```

字段说明：

| 字段 | 类型 | M1 规则 |
|---|---|---|
| `model` | string | 必填，M1 使用 `fake-seedance-lite` / `fake-seedance-pro` |
| `prompt` | string | 必填 |
| `input_image_url` | string/null | 可选，预留支持图生视频 |
| `duration` | number | 可选，默认 5，支持 5 / 10 |
| `aspect_ratio` | string | 可选，默认 `16:9`，支持 `16:9` / `9:16` / `1:1` |
| `resolution` | string | 可选，默认 `720p` |
| `test_mode` | string | M1 fake provider 专用，支持 `success` / `fail` / `slow` |

响应体：

```json
{
  "id": "task_abc123",
  "object": "video.task",
  "status": "queued",
  "model": "fake-seedance-lite",
  "provider": "fake",
  "estimated_quota": 0,
  "billing_status": "pending",
  "created_at": 1779451200,
  "updated_at": 1779451200
}
```

### 查询视频任务

```text
GET /v1/videos/tasks/:id
Authorization: Bearer <token>
```

成功响应示例：

```json
{
  "id": "task_abc123",
  "object": "video.task",
  "status": "succeeded",
  "model": "fake-seedance-lite",
  "prompt": "一只机械手在桌面拼装芯片",
  "provider": "fake",
  "provider_job_id": "fake_job_abc123",
  "estimated_quota": 0,
  "billing_status": "pending",
  "result": {
    "video_url": "https://example.com/fake-video.mp4",
    "cover_url": "https://example.com/fake-cover.jpg",
    "duration": 5,
    "metadata": {
      "fake": true
    }
  },
  "error": null,
  "created_at": 1779451200,
  "started_at": 1779451202,
  "completed_at": 1779451208,
  "updated_at": 1779451208
}
```

失败响应仍返回任务对象，`status=failed`：

```json
{
  "id": "task_failed123",
  "object": "video.task",
  "status": "failed",
  "error": {
    "code": "fake_upstream_error",
    "message": "Fake provider simulated failure"
  }
}
```

### 取消视频任务

```text
POST /v1/videos/tasks/:id/cancel
Authorization: Bearer <token>
```

规则：

- `queued` 和 `running` 可取消。
- `succeeded`、`failed`、`canceled` 是终态，不可取消。
- 取消成功后状态为 `canceled`。
- fake worker 不得继续推进已取消任务。

响应体返回更新后的任务对象。

## 状态机

M1 状态：

```text
queued
running
succeeded
failed
canceled
```

允许迁移：

```text
none -> queued
queued -> running
running -> succeeded
queued -> failed
running -> failed
queued -> canceled
running -> canceled
```

终态：

```text
succeeded
failed
canceled
```

终态任务不允许再变更。这个限制比“尽量修补状态”更重要，因为后台排障依赖状态可信。

## Fake Provider

fake provider 是 M1 的假上游，用来证明异步任务骨架成立。

行为：

- `test_mode=success`：任务从 `queued` 进入 `running`，再进入 `succeeded`，写入 fake result。
- `test_mode=fail`：任务从 `queued` 进入 `running`，再进入 `failed`，写入 `fake_upstream_error`。
- `test_mode=slow`：任务停留在 `queued` 或 `running` 足够长时间，便于测试取消。

fake provider 写入：

- `provider=fake`
- `provider_job_id=fake_job_<task suffix>`
- `provider_status`
- `provider_request_json`
- `provider_response_json`

## 后台原型

### 导航

后台 Header 增加管理员菜单：

```text
视频任务
```

普通用户不显示该入口。管理员和 root 可访问。

### 任务列表页

路径：

```text
/video-task
```

筛选：

- 状态
- 模型
- 用户 ID
- 任务 ID
- 时间范围

列表字段：

| 字段 | 说明 |
|---|---|
| 任务 ID | `task_...`，可复制 |
| 状态 | badge 展示 |
| 模型 | `fake-seedance-lite` / `fake-seedance-pro` |
| 用户 | 用户名或用户 ID |
| Token | token 名称或 ID |
| Provider | M1 为 `fake` |
| 预计额度 | M1 为 0 |
| 创建时间 | 本地时间 |
| 操作 | 查看；`queued/running` 可取消 |

顶部操作：

```text
新建调试任务
```

### 任务详情页

路径：

```text
/video-task/:id
```

模块：

- 当前状态摘要：状态、错误、结果。
- 基础信息：`task_id`、`user_id`、`token_id`、`request_id`。
- 请求参数：model、prompt、duration、aspect_ratio、resolution、input_image_url。
- Provider 信息：provider、provider_job_id、provider_status、request/response snapshot。
- 事件时间线：来自 `video_task_events`。
- 操作区：复制 `task_id`、复制查询命令、取消任务。

详情页不做视频播放器。成功结果只展示 fake URL 和 metadata。

### 调试创建表单

入口：

```text
任务列表页 -> 新建调试任务
```

字段：

- prompt
- model
- duration
- aspect_ratio
- resolution
- test_mode

后台调试表单使用当前管理员会话创建任务，任务归属当前管理员用户。创建成功后跳转任务详情页。

## 数据模型

M1 使用两张表：

```text
video_tasks
video_task_events
```

不在 M1 建 `provider_jobs`。原因是 M1 只有 fake provider、单次执行、无重试、无多上游；单独拆 `provider_jobs` 的成本高于收益。M3 接真实 Seedance，如果出现重试、多 provider 尝试或需要独立上游审计，再从 `video_tasks.provider_*` 字段拆出 `provider_jobs`。

### video_tasks

职责：保存任务当前事实，供 API 查询和后台列表使用。

建议字段：

| 字段 | 说明 |
|---|---|
| `id` | 数据库自增 ID |
| `task_id` | 对外任务 ID，唯一 |
| `user_id` | 用户 ID |
| `token_id` | Token ID |
| `channel_id` | 预留，M1 可为空或 fake |
| `request_id` | 请求 ID |
| `model` | 模型名 |
| `prompt` | 文生视频 prompt |
| `input_image_url` | 图生视频预留字段 |
| `params_json` | duration、aspect_ratio、resolution、test_mode 等 |
| `status` | 当前任务状态 |
| `provider` | M1 为 `fake` |
| `provider_job_id` | fake job ID |
| `provider_status` | fake provider 当前状态 |
| `provider_request_json` | fake provider 请求快照 |
| `provider_response_json` | fake provider 响应快照 |
| `estimated_quota` | M1 为 0 |
| `billing_status` | M1 为 `pending` |
| `result_json` | 成功结果 |
| `error_code` | 失败错误码 |
| `error_message` | 失败错误信息 |
| `created_at` | 创建时间 |
| `updated_at` | 更新时间 |
| `started_at` | 开始时间 |
| `completed_at` | 完成时间 |
| `canceled_at` | 取消时间 |

### video_task_events

职责：记录任务过程，供详情页时间线和排障使用。

建议字段：

| 字段 | 说明 |
|---|---|
| `id` | 数据库自增 ID |
| `task_id` | 对外任务 ID |
| `event_type` | `task.created` / `task.started` / `task.succeeded` / `task.failed` / `task.canceled` |
| `from_status` | 变更前状态 |
| `to_status` | 变更后状态 |
| `message` | 事件说明 |
| `metadata_json` | provider 快照、错误细节、操作人等 |
| `created_at` | 事件时间 |

## 权限

API 权限：

- 使用现有 `TokenAuth`。
- 创建、查询、取消都只能作用于当前 token 所属用户的任务。
- 管理员通过后台 API 可查看全部任务。

后台权限：

- 视频任务列表、详情、调试创建表单仅管理员可见。
- 取消任务操作仅管理员可在后台执行。
- 未来普通用户查询自己的任务，可以基于 API 权限扩展。

## 错误语义

M1 最小错误码：

| 错误码 | 场景 |
|---|---|
| `invalid_video_task_request` | 请求参数非法 |
| `video_task_not_found` | 任务不存在或无权限 |
| `video_task_not_cancelable` | 终态任务不可取消 |
| `fake_upstream_error` | fake provider 模拟失败 |
| `video_task_state_conflict` | 状态迁移冲突 |

API 错误格式复用现有 OpenAI 风格错误结构。

## 验收标准

M1 验收必须跑通 4 条链路：

1. API 创建 `success` 任务，任务从 `queued -> running -> succeeded`，查询接口返回 fake result。
2. API 创建 `fail` 任务，任务从 `queued -> running -> failed`，查询接口返回 error code/message。
3. API 创建 `slow` 任务后取消，任务进入 `canceled`，fake worker 不再推进。
4. 后台调试创建任务，创建后跳转详情页，详情页能看到状态、参数、provider 信息和事件时间线。

后台验收：

- 列表页可筛选状态、模型、用户和任务 ID。
- `queued/running` 任务显示取消按钮。
- 终态任务不显示取消按钮。
- 详情页可复制 `task_id` 和查询命令。
- 每次状态变化都有 `video_task_events` 记录。

## 后续演进

M2：

- 增加 `ledger_entries`。
- 将 `estimated_quota` 和 `billing_status` 接入预扣、结算、退款。

M3：

- 接真实 Seedance provider。
- 如出现重试、多上游或更强审计需要，从 `video_tasks.provider_*` 拆出 `provider_jobs`。

M4：

- 做普通用户 Playground。
- 增加 prompt 表单、任务进度、结果预览和下载体验。

M5：

- 增加对象存储转存。
- 增加 webhook 通知和投递记录。
