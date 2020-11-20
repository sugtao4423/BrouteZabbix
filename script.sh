#!/bin/bash

###### Watch Script Settings ######
dir="`dirname $0`"
script="$dir/broute.py"
scriptoption='PYTHONIOENCODING=utf-8'
memdir="$dir/shm"
logfile="$memdir/power.log"
###################################

######## Zabbix Settings ########
zabbix_sender='/usr/bin/zabbix_sender'
zabbix_server='ZABBIX_SERVER'
hostname='HOSTNAME'
zabbix_key='KEY'
#################################

start(){
  mkdir -p "$memdir"
  sudo mount -t tmpfs -o size=4m tmpfs "$memdir"
  sudo "$scriptoption" \
    stdbuf -i0 -o0 -e0 \
    "$script" > "$logfile" &
  echo 'Start watch B route'
}

stop(){
  sudo pkill -f "$script"
  sudo rm "$logfile"
  sudo umount "$memdir"
  sudo rm -r "$memdir"
  echo 'Stop watch B route'
}

send(){
  powerW=`timeout 3s \
    tail -f -n 0 "$logfile" | \
    grep -m 1 '瞬時電力計測値' | \
    awk -F '[:\[]' '{print $2}'`
  if [[ "$powerW" = '' ]]; then
    stop; start
    exit
  fi
  $zabbix_sender \
    -z "$zabbix_server" \
    -s "$hostname" \
    -k "$zabbix_key" \
    -o "$powerW"
  echo '' > "$logfile"
  pkill -f "$0 send"
}

if [[ "$1" = 'start' ]]; then
  start
elif [[ "$1" = 'stop' ]]; then
  stop
elif [[ "$1" = 'send' ]]; then
  send
else
  echo 'nothing'
fi
