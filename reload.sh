#!/bin/bash

# 检查是否传入了必要的参数
if [ -z "$1" ] || [ -z "$2" ]; then
    echo "Usage: $0 <binary_name> <watch_directory>"
    exit 1
fi

APP_NAME="$1"                      # 要监控的二进制文件名（如 demo）
WATCH_PATH="$2"                    # 监控的目录（如 /app）
APP_PATH="${WATCH_PATH}/${APP_NAME}"         # 上传的新文件路径
CURRENT_PATH="${WATCH_PATH}/${APP_NAME}.current"  # 运行中的文件路径


# 启动应用程序
start_app() {
    echo "Starting ${APP_NAME}..."
    mv "$APP_PATH" "$CURRENT_PATH"        # 将新上传的文件重命名
    chmod +x "$CURRENT_PATH"              # 确保文件可执行
    sleep 1
    "$CURRENT_PATH" &
    APP_PID=$!
    echo "${APP_NAME} started with PID: $APP_PID"
}

# 停止应用程序
stop_app() {
    if [ -n "$APP_PID" ]; then
        echo "Stopping ${APP_NAME} with PID: $APP_PID"
        kill "$APP_PID"
        wait "$APP_PID" 2>/dev/null
        echo "${APP_NAME} stopped"
    fi
}

# 初次启动程序
start_app

# 使用 inotifywait 监控 APP_PATH 的 close_write 事件，确保只在新的文件写入完成后触发
while true; do
    # 监听 close_write 事件，排除文件重命名的误触发
    inotifywait -e close_write --exclude "${APP_NAME}.current" "$WATCH_PATH"
    sleep 1
    # 检查是否是新的文件写入完成
    if [ -f "$APP_PATH" ]; then
        echo "Detected new version of ${APP_NAME}, restarting..."
        # 删除旧文件并启动新程序
        rm -f "$CURRENT_PATH"
        stop_app
        sleep 1
        start_app
    fi
done
