#!/usr/bin/env bash
# This script will run inside ubuntu-pod that is located at default namespace in the cluster
# This script will not and should not be executed in the self hosted runner

echo "Verify Data in remaining pool(s) after decommission test"

echo "Get data and verify files are still present as they were uploaded"
FILE_TO_SAVE_MC_OUTPUT=tmp.log
rm -rf $FILE_TO_SAVE_MC_OUTPUT
mc ls myminio/bucket > $FILE_TO_SAVE_MC_OUTPUT 2>&1
RESULT=$(grep -c ERROR $FILE_TO_SAVE_MC_OUTPUT | awk -F' ' '{print $1}')
while [ "$RESULT" == "1" ]; do
	echo "There was an error in mc ls, retrying"
	mc ls myminio/bucket > $FILE_TO_SAVE_MC_OUTPUT 2>&1
	RESULT=$(grep -c ERROR $FILE_TO_SAVE_MC_OUTPUT | awk -F' ' '{print $1}')
	echo "${RESULT}"
done
echo "mc ls was successful"
mc ls myminio/bucket # To see what is listed

for i in {1..10}; do
	RESULT=$(grep -c file"$i" "${FILE_TO_SAVE_MC_OUTPUT}")
	if [ "${RESULT}" -eq 0 ]; then
		echo "fail, there is a missing file: file${i}"
		echo "exiting with 1 as there is missing file..."
		echo "script failed" > verify-data.log
		exit 1
	fi
done
echo "script passed" > verify-data.log
