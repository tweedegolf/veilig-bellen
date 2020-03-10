#!/bin/bash

go run $1 &
while inotifywait --exclude .swp -e modify -r . ;
do
    # find PID of the file generated by `go run .` to kill it. make sure the grep does not match other processes running on the system
    IDS=$(ps ax | grep "/tmp/go-build" | grep "b001/exe/main" | grep -v "grep" | awk '{print $1}')
    if [ ! -z "$IDS" ]
    then
        kill $IDS;
    fi
    go run $1 &
done;