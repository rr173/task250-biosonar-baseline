基于 Go 实现的生物声呐航线回波底质分类 Web 项目，一款后端服务，完成声呐回波姿态几何校正、海底底质高斯似然分类与沿测线空间连续段融合。

# BENZHI 评测说明

本项目为纯后端 Go 服务，对外暴露 `/api` 前缀的 HTTP 接口，使用 SQLite 持久化，
支持进程关闭后重新打开同一数据库恢复全部业务数据。

## 标准命令

```bash
CGO_ENABLED=0 GOTOOLCHAIN=local go build ./...
CGO_ENABLED=0 GOTOOLCHAIN=local go vet   ./...
CGO_ENABLED=0 GOTOOLCHAIN=local go test  ./...
go run ./cmd/biosonar --smoke-test
go run ./cmd/biosonar --addr :8080 --db biosonar.db
```

- `--addr`：HTTP 监听地址，默认 `:8080`
- `--db`：SQLite 数据库文件路径，默认 `biosonar.db`
- `--smoke-test`：不常驻；跑完端到端场景后关闭并重新打开数据库，退出码 0 表示通过

## 冒烟自测契约（--smoke-test）

创建临时数据库 → 写入测线批次与多频回波 → 姿态几何校正 → 多频特征提取 →
底质高斯似然分类 → 沿测线连续段融合 → 发布解释快照 → 走完批次状态机至封存 →
关闭并重新打开数据库，校验批次/回波/段/快照全部持久化恢复后退出 0。

## Docker 构建与双架构验证

`Dockerfile` 与 `benzhi.Dockerfile` 内容完全一致。使用项目提供的
`build_benzhi_docker.sh` 构建评测镜像；Dockerfile 不声明端口，服务监听地址由
运行参数 `--addr` 指定。

```bash
./build_benzhi_docker.sh task250-biosonar:amd64 linux/amd64
docker run --rm task250-biosonar:amd64 --smoke-test

./build_benzhi_docker.sh task250-biosonar:arm64 linux/arm64
docker run --rm task250-biosonar:arm64 --smoke-test

docker run --rm -P task250-biosonar:amd64 --addr :8080 --db ./app.db
```

## 核心 API（`/api` 前缀）

- `POST /api/batches` 创建测线批次；`GET /api/batches`、`GET /api/batches/{id}`
- `POST /api/batches/{id}/seal` 封存批次（终态，不可变）
- `POST /api/echoes` 摄入回波窗（姿态超限自动标为异常）
- `GET /api/echoes?batch_id=` 列出回波；`GET /api/echoes/{id}`
- `POST /api/echoes/{id}/correct` 几何校正；`POST /api/echoes/{id}/exclude` 排除
- `GET /api/echoes/{id}/features` 特征；`POST|GET /api/echoes/{id}/classify` 分类
- `POST /api/substrates` / `GET /api/substrates` 底质类型目录
- `POST /api/segments/merge` 融合分段；`GET /api/segments?batch_id=`
- `POST /api/segments/{id}/confirm|reject` 确认/否决段
- `POST /api/snapshots` 发布解释快照；`GET /api/snapshots?batch_id=`、`GET /api/snapshots/{id}`
- `GET /api/stats`、`GET /api/health` 自省

## 业务不变量

- 批次状态机：receiving→pending_correction→pending_classification→published→sealed（封存后不可变）。
- 回波状态机：raw→corrected→(attitude_anomaly|excluded)；姿态超限自动标异常。
- 底质段状态机：candidate→continuous|uncertain→confirmed|rejected。
- 解释快照：draft→published→superseded（同批次仅一个 published）。
- 摄入回波拒绝时间倒退、重复 ping 序号、非法姿态/零坐标。
