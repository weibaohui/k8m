# 数据库配置说明

K8M 支持多种数据库后端，包括 SQLite、MySQL、PostgreSQL。   
## 1. 选择数据库类型

- `sqlite`：适合本地开发和轻量级部署。
- `mysql`：适合生产环境，支持高并发和大数据量。
- `postgresql`：适合对事务和一致性有更高要求的场景。

通过如下方式指定数据库类型：

- 环境变量：
  ```env
  DB_DRIVER=sqlite  # 或 mysql、postgresql
  ```
- 启动参数：
  ```shell
  --db-driver=sqlite  # 或 mysql、postgresql
  ```

---

## 2. SQLite 配置

### 基础配置

可以通过以下方式配置 SQLite：

1. **基本配置** - 仅指定数据库文件路径：
   - 环境变量：
     ```env
     SQLITE_PATH=./data/k8m.db
     ```
   - 启动参数：
     ```shell
     --sqlite-path=./data/k8m.db
     ```

2. **高级配置** - 自定义完整的 DSN 参数：
   - 环境变量：
     ```env
     SQLITE_DSN="file:./data/k8m.db?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)"
     ```
   - 启动参数：
     ```shell
     --sqlite-dsn="file:./data/k8m.db?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)"
     ```

   > **优先级说明**：
   > - 如果设置了 `SQLITE_DSN`，将优先使用此配置
   > - 如果同时设置了 `SQLITE_DSN` 和 `SQLITE_PATH`，系统会给出警告并优先使用 `SQLITE_DSN`
   > - 如果都未设置，将使用默认的优化配置
   > - 更多DSN参数，请参考https://github.com/glebarez/go-sqlite?tab=readme-ov-file#connection-string-examples
### 并发写入优化

为了降低多并发写入时出现 "database is locked" 的风险，K8M 已在内部启用了以下优化：

1. WAL (Write-Ahead Logging) 模式
2. busy_timeout 设置为 5000ms

### 注意事项

1. **目录权限**: 启动前请确保数据库文件所在目录 (如 `./data`) 已存在且具备正确的读写权限。
2. **并发限制**: SQLite 适合轻量级应用场景。如果您的应用场景涉及大量并发写入，建议考虑使用 MySQL 或 PostgreSQL。
3. **数据备份**: 由于启用了 WAL 模式，请使用以下 SQLite 命令进行安全备份：
   ```shell
   # 方法一：使用 .backup 命令（推荐）
   sqlite3 ./data/k8m.db '.backup ./data/k8m.db.backup'

   # 方法二：使用 VACUUM INTO 命令（SQLite 3.27.0 及以上版本）
   sqlite3 ./data/k8m.db 'VACUUM INTO "./data/k8m.db.backup"'
   ```
   > 注意：在 WAL 模式下，不建议直接使用 `cp` 命令复制数据库文件，这可能导致备份不完整或不一致。如果必须使用 `cp`，需要：
   > 1. 暂停所有数据库写入操作
   > 2. 同时备份 `-wal` 和 `-shm` 文件
   > 3. 确保备份期间没有新的写入
   > 
   > 因风险较高，强烈建议使用上述 SQLite 原生备份命令。

---

## 3. MySQL 配置

支持如下参数：

| 环境变量        | 启动参数          | 说明     | 默认值                   |
| --------------- | ----------------- | -------- | ------------------------ |
| MYSQL_HOST      | --mysql-host      | 主机     | 127.0.0.1                |
| MYSQL_PORT      | --mysql-port      | 端口     | 3306                     |
| MYSQL_USER      | --mysql-user      | 用户名   | root                     |
| MYSQL_PASSWORD  | --mysql-password  | 密码     | ""                       |
| MYSQL_DATABASE  | --mysql-database  | 数据库名 | k8m                      |
| MYSQL_CHARSET   | --mysql-charset   | 字符集   | utf8mb4                  |
| MYSQL_COLLATION | --mysql-collation | 排序规则 | utf8mb4_general_ci       |
| MYSQL_QUERY     | --mysql-query     | 额外参数 | parseTime=True&loc=Local |
| MYSQL_LOGMODE   | --mysql-logmode   | 日志模式 | false                    |

示例：
```env
DB_DRIVER=mysql
MYSQL_HOST=127.0.0.1
MYSQL_PORT=3306
MYSQL_USER=root
MYSQL_PASSWORD=yourpassword
MYSQL_DATABASE=k8m
MYSQL_CHARSET=utf8mb4
MYSQL_COLLATION=utf8mb4_general_ci
MYSQL_QUERY=parseTime=True&loc=Local
MYSQL_LOGMODE=false
```

---

## 4. PostgreSQL 配置

支持如下参数：

| 环境变量    | 启动参数      | 说明     | 默认值        |
| ----------- | ------------- | -------- | ------------- |
| PG_HOST     | --pg-host     | 主机     | 127.0.0.1     |
| PG_PORT     | --pg-port     | 端口     | 5432          |
| PG_USER     | --pg-user     | 用户名   | postgres      |
| PG_PASSWORD | --pg-password | 密码     | ""            |
| PG_DATABASE | --pg-database | 数据库名 | k8m           |
| PG_SSLMODE  | --pg-sslmode  | SSL模式  | disable       |
| PG_TIMEZONE | --pg-timezone | 时区     | Asia/Shanghai |
| PG_LOGMODE  | --pg-logmode  | 日志模式 | false         |

示例：
```env
DB_DRIVER=postgresql
PG_HOST=127.0.0.1
PG_PORT=5432
PG_USER=postgres
PG_PASSWORD=yourpassword
PG_DATABASE=k8m
PG_SSLMODE=disable
PG_TIMEZONE=Asia/Shanghai
PG_LOGMODE=false
```

---

## 5. 配置优先级说明

1. 启动参数（如 --mysql-host）优先级最高。
2. 环境变量（如 MYSQL_HOST）次之。
3. 若未设置，使用内置默认值。

---

## 6. 其他说明

- 所有数据库参数均可通过环境变量或启动参数配置。
- 推荐生产环境使用 MySQL 或 PostgreSQL。
- 日志模式（logmode）为 true 时，SQL 语句会输出到日志，便于调试。
- 配置变更后需重启服务生效。

如需更多帮助，请参考 `.env.example` 文件或源码注释。