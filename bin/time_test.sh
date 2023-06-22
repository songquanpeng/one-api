#!/bin/bash

if [ $# -ne 3 ]; then
  echo "Usage: time_test.sh <domain> <key> <count>"
  exit 1
fi

domain=$1
key=$2
count=$3
total_time=0
times=()

for ((i=1; i<=count; i++)); do
  result=$(curl -o /dev/null -s -w %{time_total}\\n \
           https://"$domain"/v1/chat/completions \
           -H "Content-Type: application/json" \
           -H "Authorization: Bearer $key" \
           -d '{"prompt": "hi!", "max_tokens": 1}')
  echo "$result"
  total_time=$(bc <<< "$total_time + $result")
  times+=("$result")
done

average_time=$(echo "scale=4; $total_time / $count" | bc)

sum_of_squares=0
for time in "${times[@]}"; do
  difference=$(echo "scale=4; $time - $average_time" | bc)
  square=$(echo "scale=4; $difference * $difference" | bc)
  sum_of_squares=$(echo "scale=4; $sum_of_squares + $square" | bc)
done

standard_deviation=$(echo "scale=4; sqrt($sum_of_squares / $count)" | bc)

echo "Average time: $average_time±$standard_deviation"
