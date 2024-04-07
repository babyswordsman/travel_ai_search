#!/bin/bash
pids=$(ps -ef | grep "qdrant_server/qdrant" | grep -v "grep" | awk '{print $2}')
for pid in ${pids}
do
    echo "kill qdrant server pid" $pid
    kill $pid
done
