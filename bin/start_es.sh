#!/bin/bash
CURRENT_DIR=$(cd $(dirname $0);pwd)
BASE_DIR=$(cd ..;pwd)


echo $BASE_DIR
cd $BASE_DIR/es_server

su es -c 'ES_JAVA_OPTS="-Xms512m -Xmx512m" ./bin/elasticsearch -d'

ps -ef | grep "elasticsearch" | grep -v "grep" | awk '{print $2}'