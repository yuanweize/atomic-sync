<div align="center">
  <img src="docs/assets/atomic-sync-wordmark.svg" width="560" alt="Atomic Sync">

  <p><strong>基于 rclone、理解底层分支、默认安全失败的媒体归档编排器。</strong></p>

  <p>
    <a href="README.md">English</a> ·
    <a href="docs/ARCHITECTURE.md">架构</a> ·
    <a href="docs/OPERATIONS.md">运维</a> ·
    <a href="SECURITY.md">安全</a>
  </p>
</div>

---

Atomic Sync 把一个电影目录、整部剧或一个季视为不可拆分的迁移单元：先暂存、再校验、然后发布、再次校验最终目标，最后才允许清理来源。

它还会检查 mergerfs 合并视图隐藏的真实情况：两个底层分支出现同名目录，**不代表归档已经完成**。程序会比较 StorageBox 与 GD 等物理分支中每个单元的相对文件路径和大小，区分已归档、待强校验、部分归档、待归档、冲突与空目录。

## 为什么普通移动命令会拆散剧集

`rclone move --min-age 30d` 按单个文件判断年龄。后加入的字幕、海报或剧集会留下，而旧视频先被移动，最终同一个媒体目录分散在两个位置。

Atomic Sync 读取整个单元中最新文件的修改时间。稳定窗口为 30 天时，只有完整目录连续 30 天没有变化才进入候选队列。

## 安全协议

```mermaid
flowchart LR
  A[发现稳定单元] --> B[固定目标]
  B --> C[复制到隐藏暂存区]
  C --> D[校验暂存区]
  D --> E{目标已存在?}
  E -->|否| F[发布单元]
  E -->|默认策略| X[停止并保留来源]
  E -->|不可变合并| G[只补缺失文件，绝不覆盖]
  F --> H[再次校验最终目标]
  G --> H
  H --> I{移动模式?}
  I -->|否| J[完成，来源保留]
  I -->|是| K[删除已校验来源]
```

- 新任务默认 `dry-run`。
- 目标已存在时默认安全失败。
- 破坏性的移动任务禁止文件过滤器，避免未复制的排除文件被整个单元清理误删。
- `merge-immutable` 只补缺失内容，不覆盖同路径不同文件。
- 来源到暂存、来源到最终目标都必须通过校验。
- 目录目标归属写入 SQLite，重启与重试不会重新分流。
- SIGTERM 会取消并等待任务；异常退出留下的非终态记录在重启时标记失败，暂存内容保留。
- UI 支持英文/简体中文；API Token 只保存在当前浏览器标签页。
- 容器使用 UID 1000、只读根文件系统、零 capabilities 和 `no-new-privileges`。

## mergerfs 底层归档判断

宏观状态不是按“有没有同名文件夹”判断，而是按内部文件清单判断：

| 状态 | 含义 | 建议动作 |
|---|---|---|
| `archived`（已归档） | GD 有内容，来源已经没有文件（可残留空目录壳） | 确认挂载健康并保留审计记录 |
| `ready-to-verify`（待强校验/清理） | 来源的相对路径和大小全部能在 GD 对应，但来源仍在 | 删除前执行 checksum/size check |
| `partial`（部分归档） | GD 中已有部分内容，但仍缺少来源文件 | 不可变合并或人工核对 |
| `pending`（待归档） | 来源有文件，目标不存在该单元 | 归档候选 |
| `conflict`（冲突） | 相同相对路径的文件大小或文件/目录类型不同 | 立即停止，人工选择权威副本 |
| `empty`（空目录） | 只有空目录壳 | 检查或忽略 |

控制台分析只读取路径和大小，避免为了画面扫描就读取数 TB 的 CIFS 文件；任何来源删除仍需要真正的最终校验。完整规则见 [归档分析说明](docs/ARCHIVE-ANALYSIS.md)。

`archived` 是根据当前物理分支清单推断的状态，不是某次成功运行的历史证明。全部归档完成后，来源成功返回空清单是合法状态，所以分析前必须先确认物理挂载健康。参考 Compose 会拒绝自动创建缺失的 bind 来源，但无法区分“健康的空共享”和“目录仍在、底层文件系统已离线”。

## 快速开始

需要 Docker Engine、Compose v2 和现有 `rclone.conf`。

```bash
git clone https://github.com/yuanweize/atomic-sync.git
cd atomic-sync
mkdir -p data rclone source
cp /path/to/rclone.conf rclone/rclone.conf
cp .env.example .env
openssl rand -hex 32
```

把生成的 Token 写入 `.env`，然后只处理 Atomic Sync 的专用目录权限；**绝不要对共享媒体根目录递归 `chown`**。

```bash
sudo chown -R 1000:1000 data
sudo chown 1000:1000 rclone/rclone.conf
sudo chmod 600 rclone/rclone.conf
docker compose -f compose.yaml -f compose.dev.yaml up -d --build
docker compose ps
```

打开 `http://127.0.0.1:8088`，输入 API Token。第一次只创建“复制 + 仅演练”任务。默认来源挂载为容器内 `/sources/media`，并且是只读的。

## 生产部署

正式 Release 会生成带 SBOM、provenance 和 keyless 签名的 `linux/amd64`、`linux/arm64` 镜像。生产必须固定 Release digest：

```yaml
services:
  atomic-sync:
    image: ghcr.io/yuanweize/atomic-sync:0.1.0@sha256:<release-digest>
```

把服务嵌入其他 Compose 项目时不要使用 `build: .`，否则构建上下文会变成对方项目目录。

每个 GitHub Release 都包含 `image-digest.txt`、Linux 二进制文件和 `SHA256SUMS`。发布流程会把 GHCR 包设为公开、验证匿名读取、附带 SBOM/provenance，并使用 GitHub OIDC 对不可变 digest 签名。部署前可这样验证：

```bash
IMAGE="$(cat image-digest.txt)"
docker pull "$IMAGE"
cosign verify "$IMAGE" \
  --certificate-identity-regexp '^https://github.com/yuanweize/atomic-sync/.github/workflows/release.yml@refs/tags/v[0-9]+\.[0-9]+\.[0-9]+$' \
  --certificate-oidc-issuer 'https://token.actions.githubusercontent.com'
sha256sum -c SHA256SUMS
```

## 配置

| 环境变量 | 默认值 | 用途 |
|---|---:|---|
| `ATOMIC_LISTEN` | `127.0.0.1:8080` | HTTP 监听地址；容器会显式设置为 `:8080` |
| `ATOMIC_DATA_DIR` | `/data` | SQLite 与持久化状态目录 |
| `ATOMIC_API_TOKEN` | 空 | Bearer Token；设置后至少 32 个字符，非 loopback 监听必须配置 |
| `ATOMIC_RCLONE_BIN` | `rclone` | rclone 可执行文件 |
| `ATOMIC_MAX_CONCURRENCY` | `4` | 全局传输并发上限 |
| `ATOMIC_LOG_FORMAT` | `json` | `json` 或文本结构化日志 |
| `RCLONE_CONFIG` | `/config/rclone/rclone.conf` | 明确的 rclone 配置路径 |

程序不会修改 `.env` 或 `rclone.conf`。任务、目标归属、运行历史与分支分析结果均保存在 `atomic-sync.db`。

## 真实保证边界

对象存储的“目录移动”不是真正的 ACID 事务，远程 `moveto` 可能由多个对象操作组成。失败时 Atomic Sync 保留来源和暂存内容，便于诊断与重试。

`merge-immutable` 也不是完全原子的：发现后续冲突前，部分缺失对象可能已经补到目标；但它不会覆盖不同的目标对象，也不会在最终校验前删除来源。

SQLite 架构只支持一个 Atomic Sync 实例，不要让多个副本共享同一个数据库。

## 文档

- [架构与状态机](docs/ARCHITECTURE.md)
- [mergerfs 分支归档分析](docs/ARCHIVE-ANALYSIS.md)
- [生产运维、升级与回滚](docs/OPERATIONS.md)
- [HTTP API](docs/API.md)
- [威胁模型](docs/SECURITY-MODEL.md)
- [贡献指南](CONTRIBUTING.md)
- [安全策略](SECURITY.md)

## 开发验证

需要 Go 1.25 或更高版本。

```bash
gofmt -w .
go vet ./...
go test -race -cover ./...
ATOMIC_API_TOKEN="$(openssl rand -hex 32)" docker compose config
docker build -t atomic-sync:dev .
```

测试覆盖 dry-run、安全冲突、不可变合并、关机恢复、鉴权、API、SQLite，以及 mergerfs 同名目录内部文件部分归档的判断。

## 路线图

- 定时轮询与文件系统监听适配器
- 对选定分析单元按需计算校验和
- 暂存内容续传与引导式清理
- Prometheus 指标、通知与细粒度 RBAC
- 面向远程 NAS 推送的多节点 Agent

项目使用 [MIT License](LICENSE)。
