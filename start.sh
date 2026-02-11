#!/bin/bash
cd "$(dirname "$0")"

if [ -f .pid ] && kill -0 "$(cat .pid)" 2>/dev/null; then
  echo "Server is already running (PID: $(cat .pid))"
  exit 1
fi

go build -o playcamp-go-sdk-example . || exit 1

./playcamp-go-sdk-example &
echo $! > .pid
echo "Server started (PID: $!)"
