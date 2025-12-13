#!/bin/bash

# Start each service in background
go run cmd/auth/main.go &
go run cmd/file/main.go &
go run cmd/channel/main.go &
go run cmd/payment/main.go &
# go run cmd/gateway/main.go &


# Wait to keep script running
wait


# cmd for running
# first :chmod +x scripts/run_all.sh
# then : ./scripts/run_all.sh


