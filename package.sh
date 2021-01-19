#!/bin/bash

binary=$(basename "$(dirname "${1}")")
if [[ "${binary}" =~ kubectl-minio* ]]; then
    cp -f LICENSE "$(dirname "${1}")"
    cp -f CREDITS "$(dirname "${1}")"
    cp -f README.md  "$(dirname "${1}")"
    zip -q -r -j "${binary}.zip" "$(dirname "${1}")"
fi
