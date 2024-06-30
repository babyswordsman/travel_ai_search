#!/bin/bash
pids=$(ps -ef | grep "quickwit/quickwit" | grep -v "grep" | awk '{print $2}')
for pid in ${pids}
do
    echo "kill quickwit server pid" $pid
    kill $pid
done
