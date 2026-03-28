#!/bin/bash
[[ -z "$1" ]] && echo "Error: argument required" && exit 1
bluetoothctl "$@"
