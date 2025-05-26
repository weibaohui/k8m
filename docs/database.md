# 数据库配置说明

K8M 支持多种数据库后端，包括 SQLite、MySQL 和 PostgreSQL。数据库类型通过 `DB_DRIVER` 环境变量或启动参数指定。

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

仅需指定数据库文件路径：

- 环境变量：
  ```env
  SQLITE_PATH=./data/k8m.db
  ```
- 启动参数：
  ```shell
  --sqlite-path=./data/k8m.db
  ```

---

## 3. MySQL 配置

支持如下参数：

| 环境变量         | 启动参数             | 说明           | 默认值                |
|------------------|----------------------|----------------|-----------------------|
| MYSQL_HOST       | --mysql-host         | 主机           | 127.0.0.1             |
| MYSQL_PORT       | --mysql-port         | 端口           | 3306                  |
| MYSQL_USER       | --mysql-user         | 用户名         | root                  |
| MYSQL_PASSWORD   | --mysql-password     | 密码           | ""                    |
| MYSQL_DATABASE   | --mysql-database     | 数据库名       | k8m                   |
| MYSQL_CHARSET    | --mysql-charset      | 字符集         | utf8mb4               |
| MYSQL_COLLATION  | --mysql-collation    | 排序规则       | utf8mb4_general_ci    |
| MYSQL_QUERY      | --mysql-query        | 额外参数       | parseTime=True&loc=Local |
| MYSQL_LOGMODE    | --mysql-logmode      | 日志模式       | false                 |

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

| 环境变量         | 启动参数             | 说明           | 默认值                |
|------------------|----------------------|----------------|-----------------------|
| PG_HOST          | --pg-host            | 主机           | 127.0.0.1             |
| PG_PORT          | --pg-port            | 端口           | 5432                  |
| PG_USER          | --pg-user            | 用户名         | postgres              |
| PG_PASSWORD      | --pg-password        | 密码           | ""                    |
| PG_DATABASE      | --pg-database        | 数据库名       | k8m                   |
| PG_SSLMODE       | --pg-sslmode         | SSL模式        | disable               |
| PG_TIMEZONE      | --pg-timezone        | 时区           | Asia/Shanghai         |
| PG_LOGMODE       | --pg-logmode         | 日志模式       | false                 |

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