#!/bin/bash
trap "kill 0" EXIT

go build -o grassdb main.go

echo "Starting Node 1..."
./grassdb -id node1 -addr :50051 -http :8081 -peers :50052,:50053 > node1.log 2>&1 &
PID1=$!

echo "Starting Node 2..."
./grassdb -id node2 -addr :50052 -http :8082 -peers :50051,:50053 > node2.log 2>&1 &
PID2=$!

echo "Starting Node 3..."
./grassdb -id node3 -addr :50053 -http :8083 -peers :50051,:50052 > node3.log 2>&1 &
PID3=$!

echo "Cluster started. PIDs: $PID1, $PID2, $PID3"
echo "Waiting for leader election..."
sleep 5

echo "Checking logs for leader..."
grep "state=Leader" node*.log || echo "No explicit leader log found yet (maybe check internal logs)"

# Run for 15 seconds
sleep 300
echo "Test finished. Stopping nodes..."
