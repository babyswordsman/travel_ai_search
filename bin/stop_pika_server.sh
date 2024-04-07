#!/bin/bash
pids=$(ps -ef | grep "pika_server/pika" | grep -v "grep" | awk '{print $2}')
for pid in ${pids}
do
    echo "kill pika server pid" $pid
    kill $pid
done
