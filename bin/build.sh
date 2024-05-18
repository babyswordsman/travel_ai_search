#!/bin/bash
# source /etc/profile.d/go.sh
CURRENT_DIR=$(cd $(dirname $0);pwd)
BASE_DIR=$(cd ..;pwd)
echo $BASE_DIR
cd $BASE_DIR/search
go build -o ../ai_search_server server/main.go

