#!/bin/bash
CURRENT_DIR=$(cd $(dirname $0);pwd)
BASE_DIR=$(cd ..;pwd)


echo $BASE_DIR
cd $BASE_DIR/es_server

cur_user=$(whoami)

if [ "$cur_user" == "root" ]; then
    su es -c 'ES_JAVA_OPTS="-Xms512m -Xmx512m" ./bin/elasticsearch -d'
else
    ES_JAVA_OPTS="-Xms512m -Xmx512m" $BASE_DIR/es_server/bin/elasticsearch -d
fi

ps -ef | grep "elasticsearch" | grep -v "grep" | awk '{print $2}'