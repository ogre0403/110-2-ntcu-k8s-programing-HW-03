#!/bin/bash

DEADLINE='2022-06-24 00:10:00'

CURRENT=`TZ=Asia/Taipei date +"%Y-%m-%d %H:%M:%S"`


# echo "deadline: $DEADLINE"
# echo "current : $CURRENT"



if [[ $CURRENT < $DEADLINE ]]; then
    exit 0
else
    echo "deadline為 $DEADLINE, 送出時間為 $CURRENT, 已遲交"
    exit 1
fi


