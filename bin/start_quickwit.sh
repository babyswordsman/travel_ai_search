#!/bin/bash
CURRENT_DIR=$(cd $(dirname $0);pwd)
BASE_DIR=$(cd ..;pwd)

mkdir -p "$BASE_DIR/logs"
echo $BASE_DIR
cd $BASE_DIR/quickwit
nohup $BASE_DIR/quickwit/quickwit run >>  $BASE_DIR/logs/quickwit.log 2>&1 &
ps -ef | grep "quickwit/quickwit" | grep -v "grep" | awk '{print $2}'