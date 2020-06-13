#!/bin/sh

set -e

if tmux has-session -t service 2> /dev/null; then
  tmux attach -t service
  exit
fi

tmux new-session -d -s service -n editor
tmux send-keys -t service:editor "v " Enter
tmux split-window -t service:editor -h -p 10
tmux attach -t service:editor.top
