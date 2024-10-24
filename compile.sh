#!/bin/bash

# 项目名称，可以根据需要更改
PROJECT_NAME="ip-location"

# 源代码目录（当前目录）
SRC_DIR="."

# 输出目录
OUTPUT_DIR="output"

# 创建输出目录
mkdir -p $OUTPUT_DIR

# 获取所有Go语言支持的GOOS/GOARCH组合
PLATFORMS=$(go tool dist list)

# 遍历所有目标平台进行交叉编译
for PLATFORM in $PLATFORMS; do
    # 分割GOOS和GOARCH
    GOOS=${PLATFORM%/*}
    GOARCH=${PLATFORM#*/}

    # 设置输出文件名，Windows系统添加.exe后缀
    OUTPUT_FILE="$OUTPUT_DIR/${PROJECT_NAME}_${GOOS}_${GOARCH}"
    if [ "$GOOS" = "windows" ]; then
        OUTPUT_FILE="${OUTPUT_FILE}.exe"
    fi

    # 编译，使用 -ldflags 来减小文件体积
    echo "编译 $GOOS/$GOARCH..."
    env GOOS=$GOOS GOARCH=$GOARCH go build -ldflags="-s -w" -o "$OUTPUT_FILE" "$SRC_DIR"
    
    # 如果编译失败，则跳过该平台
    if [ $? -ne 0 ]; then
        echo "编译 $GOOS/$GOARCH 失败，跳过..."
        continue
    fi

    # 压缩为 tar.gz
    echo "压缩 $OUTPUT_FILE..."
    tar -czvf "$OUTPUT_FILE.tar.gz" -C "$OUTPUT_DIR" "$(basename $OUTPUT_FILE)"
    
    # 如果压缩失败，跳过该平台
    if [ $? -ne 0 ]; then
        echo "压缩 $OUTPUT_FILE 失败，跳过..."
        continue
    fi

    # 删除未压缩的二进制文件
    rm "$OUTPUT_FILE"
done

echo "所有平台的编译和压缩已完成"
