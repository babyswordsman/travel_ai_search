#!/bin/bash
CURRENT_DIR=$(cd $(dirname $0);pwd)
BASE_DIR=$(cd ..;pwd)

mkdir -p "$BASE_DIR/logs"
echo $BASE_DIR
cd $BASE_DIR
nohup $BASE_DIR/ai_search_server -conf $BASE_DIR/config/conf.yaml >>  $BASE_DIR/logs/ai_search.log 2>&1 &
ps -ef | grep "ai_search_server" | grep -v "grep" | awk '{print $2}'