#!/bin/bash
CURRENT_DIR=$(cd $(dirname $0);pwd)
BASE_DIR=$(cd ..;pwd)

mkdir -p "$BASE_DIR/logs"
echo $BASE_DIR
cd $BASE_DIR/qdrant_server
nohup $BASE_DIR/qdrant_server/qdrant --config-path $BASE_DIR/qdrant_server/config/config.yaml  >>  $BASE_DIR/logs/qdrant_server.log 2>&1 &
ps -ef | grep "qdrant_server/qdrant" | grep -v "grep" | awk '{print $2}'