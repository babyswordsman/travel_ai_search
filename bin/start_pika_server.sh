#!/bin/bash
CURRENT_DIR=$(cd $(dirname $0);pwd)
BASE_DIR=$(cd ..;pwd)

mkdir -p "$BASE_DIR/logs"
echo $BASE_DIR
cd $BASE_DIR/pika_server
nohup $BASE_DIR/pika_server/pika -c $BASE_DIR/pika_server/conf/pika.conf >>  $BASE_DIR/logs/pika_server.log 2>&1 &
ps -ef | grep "pika_server/pika" | grep -v "grep" | awk '{print $2}'