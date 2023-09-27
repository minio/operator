#!/bin/bash

if [ -f "$NVM_DIR/nvm.sh" ]
then
    \. "$NVM_DIR/nvm.sh";
    nvm use;
fi
yarn find-deadcode
