#!/bin/bash
CURRENT_DIR=$(cd $(dirname $0);pwd)
BASE_DIR=$(cd ..;pwd)

mkdir -p "$BASE_DIR/logs"
echo $BASE_DIR
nohup python3 $BASE_DIR/model_service/model_server.py >>  $BASE_DIR/logs/model_server.log 2>&1 &
ps -ef | grep "model_server.py" | grep -v "grep" | awk '{print $2}'