default: build-static

build-static:
	@echo "Building frontend static assets to 'build'"
	@if [ -f "${NVM_DIR}/nvm.sh" ]; then \. "${NVM_DIR}/nvm.sh" && nvm install && nvm use; fi && \
	  NODE_OPTIONS=--openssl-legacy-provider yarn build

test-warnings:
	./check-warnings.sh

test-prettier:
	./check-prettier.sh

prettify:
	yarn prettier --write . --loglevel warn

pretty: prettify

find-deadcode:
	./check-deadcode.sh
