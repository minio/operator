#!/usr/bin/env bash
# This script will run inside ubuntu-pod that is located at default namespace in the cluster
# This script will not and should not be executed in the self hosted runner

LOG_NAME=upload-data.log
echo "script failed" >$LOG_NAME # assume initial state
echo "Starting uploading data to MinIO Server"

echo "Generate some data & upload it"

# Create the bucket
mc mb myminio/bucket
sleep 10
for i in {1..10}; do
	echo "content$i" | mc pipe myminio/bucket/file"$i"
done
echo "script passed" >$LOG_NAME
