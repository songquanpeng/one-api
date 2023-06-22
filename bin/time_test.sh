#!/bin/bash

if [ $# -ne 3 ]; then
  echo "Usage: time_test.sh <domain> <key> <count>"
  exit 1
fi

domain=$1
key=$2
count=$3
total_time=0

for ((i=1; i<=count; i++)); do
  result=$(curl -o /dev/null -s -w %{time_total}\\n \
           https://"$domain"/v1/chat/completions \
           -H "Content-Type: application/json" \
           -H "Authorization: Bearer $key" \
           -d '{"prompt": "hi!", "max_tokens": 1}')
  echo "$result"
  total_time=$(echo "$total_time + $result" | bc)
done

average_time=$(echo "scale=3; $total_time / $count" | bc)
echo "Average time: $average_time"

