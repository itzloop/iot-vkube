#!/usr/bin/env bash
ROOT=$(pwd)
set -x
MAX="5"
SMART_LOCK_HOST='localhost:5000'
if [[ ! -z $1 ]]; then
  SMART_LOCK_HOST="$1"
fi

if [[ ! -z $2 ]]; then
  MAX=$2
fi

# System Information
echo OS Info: `uname -a`
echo HW Info:
echo -e "  Model: `lshw -short 2>/dev/null | grep system | sed 's/system//g' |  sed -e 's/^[ \t]*//'`"
echo -e "  CPU: `nproc`"
echo -e "  Memory: `echo $(grep MemTotal /proc/meminfo | awk '{print $2 / (1024 * 1024)}') GiB`"

run() {
  t="5s"
  if [[ ! -z $3 ]]; then
    t="$3"
  fi
  $ROOT/benchmark -caddr $SMART_LOCK_HOST -controllers $1 -devices $2 -d $t
}


$ROOT/smartlock -addr :5000 -controllers 1 -devices $(echo "10 ^ $MAX" | bc) > /dev/null &

# change devices
for (( i = 0; i <= $MAX; i++ )); do
  run 1 $(echo "10 ^ $i" | bc)
done

pkill smartlock
sleep 10

$ROOT/smartlock -addr :5000 -controllers $(echo "10 ^ $MAX" | bc) -devices 1 > /dev/null &

# change controllers
for (( i = 0; i <= 3; i++ )); do
  run $(echo "10 ^ $i" | bc) 1 "5s"
done

for (( i = 4; i <= $MAX; i++ )); do
  run $(echo "10 ^ $i" | bc) 1 "60s"
done

pkill smartlock