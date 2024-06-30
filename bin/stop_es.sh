#!/bin/bash
pids=$(ps -ef | grep "elasticsearch" | grep -v "grep" | awk '{print $2}')
for pid in ${pids}
do
    echo "kill elastic server pid" $pid
    kill $pid
done
