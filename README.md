# biosonar — 生物声呐航线回波底质分类服务

纯后端 Go 服务：将多频声呐回波沿测线校正几何、提取背向散射特征、用高斯似然分类
海底底质，并沿测线融合空间连续的底质段，最终发布可引用的解释快照。数据持久化于
SQLite，支持关闭后重新打开恢复。

## 构建与运行

```bash
export CGO_ENABLED=0 GOTOOLCHAIN=local
go build ./...
go vet ./...
go test ./...
go run ./cmd/biosonar --smoke-test          # 健康自检
go run ./cmd/biosonar --addr :8080 --db biosonar.db   # 常驻服务
```

## 目录结构

```
cmd/biosonar/              # 入口（--smoke-test 契约）
internal/model/            # 实体、状态机、错误
internal/store/            # SQLite 持久化（迁移 + 各实体 CRUD）
internal/geometry/         # 姿态/声速/吃水几何校正
internal/echo/             # 多频背向散射特征提取
internal/classify/         # 底质高斯似然分类
internal/segment/          # 沿测线空间连续段融合 + 边界检测
internal/versioning/       # 解释快照发布/替代策略
internal/service/          # 编排层（跨实体不变量）
internal/httpapi/          # /api HTTP 层
go.mod / go.sum
component-versions.json
Dockerfile / benzhi.Dockerfile / build_benzhi_docker.sh
BENZHI_README.md
```

## 端到端示例

1. `POST /api/batches` 创建测线批次。
2. `POST /api/echoes` 批量摄入多频回波（姿态异常自动标记）。
3. `POST /api/echoes/{id}/classify` 逐 ping 分类（内部自动校正+特征提取）。
4. `POST /api/segments/merge` 融合连续底质段。
5. `POST /api/snapshots` 发布解释快照（自动替代同批次旧快照）。
6. `POST /api/batches/{id}/seal` 封存批次。

详见 `BENZHI_README.md`。
