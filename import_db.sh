#!/bin/bash

# 数据库连接参数
DB_USER="root"
DB_PASSWORD="tsdd123456"
DB_NAME="im"

# 数据库初始化文件路径
SQL_FILE="init_db.sql"

echo "========================================"
echo "  一键导入数据库脚本 (Docker 环境)"
echo "========================================"
echo "当前目录: $(pwd)"
echo "脚本路径: $(realpath "$0")"

# 检查 SQL 文件是否存在
echo "步骤 0: 检查 SQL 文件..."
if [ ! -f "$SQL_FILE" ]; then
    echo "错误：$SQL_FILE 文件不存在！"
    ls -la
    exit 1
fi

echo "SQL 文件存在，大小: $(du -h "$SQL_FILE" | cut -f1)"
echo "SQL 文件前 10 行:"
head -10 "$SQL_FILE"
echo ""

# 检查 Docker 是否安装
echo "步骤 1: 检查 Docker 环境..."
if ! command -v docker &> /dev/null; then
    echo "错误：Docker 未安装！"
    exit 1
fi

echo "Docker 版本: $(docker --version)"

# 检查运行中的容器
echo "运行中的容器:"
docker ps

echo ""
echo "所有容器:"
docker ps -a

echo ""

# 尝试查找 MySQL 容器
echo "步骤 2: 查找 MySQL 容器..."
MYSQL_CONTAINERS=$(docker ps --format '{{.Names}}' | grep -i mysql)

if [ -z "$MYSQL_CONTAINERS" ]; then
    echo "未找到运行中的 MySQL 容器，尝试查找所有 MySQL 相关容器..."
    MYSQL_CONTAINERS=$(docker ps -a --format '{{.Names}}' | grep -i mysql)
fi

echo "找到的 MySQL 容器: $MYSQL_CONTAINERS"

# 选择第一个容器（如果有多个）
MYSQL_CONTAINER=$(echo "$MYSQL_CONTAINERS" | head -1)

echo ""

if [ -n "$MYSQL_CONTAINER" ]; then
    echo "步骤 3: 尝试连接到容器 $MYSQL_CONTAINER..."
    
    # 检查容器状态
    CONTAINER_STATUS=$(docker inspect --format '{{.State.Status}}' "$MYSQL_CONTAINER")
    echo "容器状态: $CONTAINER_STATUS"
    
    if [ "$CONTAINER_STATUS" != "running" ]; then
        echo "警告：容器未运行，尝试启动..."
        docker start "$MYSQL_CONTAINER"
        sleep 3
    fi
    
    # 尝试执行简单的 MySQL 命令
    echo "尝试执行 MySQL 版本命令..."
    docker exec "$MYSQL_CONTAINER" mysql -u "$DB_USER" -p"$DB_PASSWORD" -e "SELECT VERSION();"
    
    if [ $? -eq 0 ]; then
        echo "MySQL 连接成功！"
        
        # 创建数据库
        echo "步骤 4: 创建数据库 $DB_NAME..."
        docker exec "$MYSQL_CONTAINER" mysql -u "$DB_USER" -p"$DB_PASSWORD" -e "CREATE DATABASE IF NOT EXISTS $DB_NAME CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
        
        if [ $? -eq 0 ]; then
            echo "数据库创建成功！"
            
            # 导入数据
            echo "步骤 5: 导入数据库结构和初始数据..."
            echo "执行命令: docker exec -i '$MYSQL_CONTAINER' mysql -u '$DB_USER' -p'$DB_PASSWORD' '$DB_NAME' < '$SQL_FILE'"
            
            # 尝试导入数据
            docker exec -i "$MYSQL_CONTAINER" mysql -u "$DB_USER" -p"$DB_PASSWORD" "$DB_NAME" < "$SQL_FILE"
            
            if [ $? -eq 0 ]; then
                echo "========================================"
                echo "  数据库导入成功！"
                echo "========================================"
                echo "数据库名称: $DB_NAME"
                echo "MySQL 容器: $MYSQL_CONTAINER"
                echo "========================================"
                exit 0
            else
                echo "错误：数据导入失败！"
                echo "尝试使用不同的命令格式..."
                
                # 尝试使用另一种方式导入
                echo "尝试使用 cat 命令导入..."
                cat "$SQL_FILE" | docker exec -i "$MYSQL_CONTAINER" mysql -u "$DB_USER" -p"$DB_PASSWORD" "$DB_NAME"
                
                if [ $? -eq 0 ]; then
                    echo "========================================"
                    echo "  数据库导入成功！"
                    echo "========================================"
                    echo "数据库名称: $DB_NAME"
                    echo "MySQL 容器: $MYSQL_CONTAINER"
                    echo "========================================"
                    exit 0
                else
                    echo "错误：所有导入尝试都失败了！"
                fi
            fi
        else
            echo "错误：创建数据库失败！"
        fi
    else
        echo "错误：MySQL 连接失败！"
        echo "尝试不使用密码文件..."
        
        # 尝试不使用密码参数（让 MySQL 提示输入密码）
        echo "尝试使用交互式密码输入..."
        echo "$DB_PASSWORD" | docker exec -i "$MYSQL_CONTAINER" mysql -u "$DB_USER" -p -e "SELECT VERSION();"
        
        if [ $? -eq 0 ]; then
            echo "交互式密码输入成功！"
            echo "现在尝试创建数据库..."
            echo "$DB_PASSWORD" | docker exec -i "$MYSQL_CONTAINER" mysql -u "$DB_USER" -p -e "CREATE DATABASE IF NOT EXISTS $DB_NAME CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
            
            if [ $? -eq 0 ]; then
                echo "数据库创建成功！"
                echo "现在尝试导入数据..."
                echo "$DB_PASSWORD" | docker exec -i "$MYSQL_CONTAINER" mysql -u "$DB_USER" -p "$DB_NAME" < "$SQL_FILE"
                
                if [ $? -eq 0 ]; then
                    echo "========================================"
                    echo "  数据库导入成功！"
                    echo "========================================"
                    echo "数据库名称: $DB_NAME"
                    echo "MySQL 容器: $MYSQL_CONTAINER"
                    echo "========================================"
                    exit 0
                fi
            fi
        fi
    fi
else
    echo "错误：未找到 MySQL 容器！"
    echo "尝试使用 docker-compose 检查..."
    if [ -f "docker-compose.yml" ]; then
        echo "找到 docker-compose.yml 文件，检查服务..."
        docker-compose ps
    fi
fi

# 所有尝试都失败
echo ""
echo "========================================"
echo "  错误：数据库导入失败！"
echo "========================================"
echo "详细诊断信息:"
echo "1. 当前用户: $(whoami)"
echo "2. Docker 权限: $(ls -la /var/run/docker.sock 2>/dev/null || echo '无法访问')"
echo "3. SQL 文件权限: $(ls -la "$SQL_FILE")"
echo "4. 可用内存: $(free -h | grep Mem)"
echo "5. 可用磁盘: $(df -h .)"
echo ""
echo "解决方案:"
echo "1. 确保 MySQL 容器正在运行: docker start <容器名称>"
echo "2. 检查容器日志: docker logs <容器名称>"
echo "3. 手动测试连接: docker exec -it <容器名称> mysql -u root -p"
echo "4. 检查环境变量: docker inspect <容器名称> | grep MYSQL_"
echo "5. 尝试手动导入:"
echo "   a. 复制 SQL 文件到容器: docker cp init_db.sql <容器名称>:/tmp/"
echo "   b. 进入容器: docker exec -it <容器名称> bash"
echo "   c. 在容器内: mysql -u root -p"
echo "   d. 创建数据库: CREATE DATABASE IF NOT EXISTS im CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
echo "   e. 使用数据库: USE im;"
echo "   f. 导入数据: source /tmp/init_db.sql;"
echo ""
echo "6. 如果使用 docker-compose, 尝试:"
echo "   docker-compose exec mysql mysql -u root -p tsdd123456 im < init_db.sql"
exit 1
