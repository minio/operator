default: build-static

build-static:
	@echo "Building frontend static assets to 'build'"
	@if [ -f "${NVM_DIR}/nvm.sh" ]; then \. "${NVM_DIR}/nvm.sh" && nvm install && nvm use; fi && \
	  yarn build

test-warnings build-static:
	./check-warnings.sh

test-prettier:
	./check-prettier.sh

prettify:
	yarn prettier --write . --loglevel warn

pretty: prettify

find-deadcode:
	./check-deadcode.sh
