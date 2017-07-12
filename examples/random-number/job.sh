#!/usr/bin/env sh

awk 'BEGIN { srand();printf("%d\n",rand()*60)  }' | tee $1/number.txt

echo $NUMBER > $1/number.txt
