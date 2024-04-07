#!/bin/bash
pids=$(ps -ef | grep "model_server.py" | grep -v "grep" | awk '{print $2}')
for pid in ${pids}
do
    echo "kill pid" $pid
    kill $pid
done
