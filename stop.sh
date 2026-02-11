#!/bin/bash
cd "$(dirname "$0")"

# Try .pid file first.
if [ -f .pid ]; then
  PID=$(cat .pid)
  if kill -0 "$PID" 2>/dev/null; then
    kill "$PID"
    echo "Server stopped (PID: $PID)"
  else
    echo "PID $PID not running"
  fi
  rm -f .pid
fi

# Also kill any remaining process.
REMAINING=$(pgrep -f "./playcamp-go-sdk-example" 2>/dev/null)
if [ -n "$REMAINING" ]; then
  kill $REMAINING 2>/dev/null
  echo "Cleaned up remaining process: $REMAINING"
fi
