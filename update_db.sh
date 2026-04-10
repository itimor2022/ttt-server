#!/bin/bash

echo "========================================"
echo "  数据库更新脚本"
echo "========================================"
echo "当前目录: $(pwd)"
echo "脚本路径: $(realpath "$0")"
echo ""

# 数据库连接参数
DB_USER="root"
DB_PASSWORD="demo"
DB_NAME="test"

# 数据库初始化文件路径
SQL_FILE="init_db.sql"

# 检查 SQL 文件是否存在
echo "步骤 0: 检查 SQL 文件..."
if [ ! -f "$SQL_FILE" ]; then
    echo "错误：$SQL_FILE 文件不存在！"
    ls -la
    exit 1
fi

echo "SQL 文件存在，大小: $(du -h "$SQL_FILE" | cut -f1)"
echo ""

# 检查 Docker 是否安装
echo "步骤 1: 检查 Docker 环境..."
if ! command -v docker &> /dev/null; then
    echo "错误：Docker 未安装！"
    exit 1
fi

echo "Docker 版本: $(docker --version)"
echo ""

# 检查 MySQL 容器
echo "步骤 2: 检查 MySQL 容器..."
if ! docker ps --format '{{.Names}}' | grep -q "tsdd-mysql-1"; then
    echo "错误：tsdd-mysql-1 容器不存在！"
    exit 1
fi

echo "MySQL 容器 tsdd-mysql-1 存在"
echo ""

# 复制 SQL 文件到容器
echo "步骤 3: 复制 SQL 文件到容器..."
docker cp "$SQL_FILE" tsdd-mysql-1:/tmp/

if [ $? -eq 0 ]; then
    echo "SQL 文件复制成功！"
else
    echo "错误：SQL 文件复制失败！"
    exit 1
fi

echo ""

# 执行数据库更新
echo "步骤 4: 执行数据库更新..."
echo "执行命令: docker exec tsdd-mysql-1 bash -c 'mysql -u $DB_USER -p$DB_PASSWORD $DB_NAME < /tmp/$SQL_FILE'"
echo ""
docker exec tsdd-mysql-1 bash -c "mysql -u $DB_USER -p$DB_PASSWORD $DB_NAME < /tmp/$SQL_FILE"

if [ $? -eq 0 ]; then
    echo ""
    echo "========================================"
    echo "  数据库更新成功！🎉"
    echo "========================================"
    echo "数据库名称: $DB_NAME"
    echo "MySQL 容器: tsdd-mysql-1"
    echo "========================================"
else
    echo ""
    echo "========================================"
    echo "  数据库更新失败！"
    echo "========================================"
    echo "请检查错误信息并重新尝试。"
    echo "========================================"
    exit 1
fi

# 验证更新结果
echo ""
echo "步骤 5: 验证更新结果..."
echo "检查用户表记录数:"
docker exec tsdd-mysql-1 bash -c "mysql -u $DB_USER -p$DB_PASSWORD $DB_NAME -e 'SELECT COUNT(*) FROM user;'"
echo ""
echo "检查钱包表结构:"
docker exec tsdd-mysql-1 bash -c "mysql -u $DB_USER -p$DB_PASSWORD $DB_NAME -e 'DESCRIBE wallet;'"
echo ""
echo "检查用户在线表结构:"
docker exec tsdd-mysql-1 bash -c "mysql -u $DB_USER -p$DB_PASSWORD $DB_NAME -e 'DESCRIBE user_online;'"

 echo ""
echo "========================================"
echo "  数据库更新完成！"
echo "========================================"
