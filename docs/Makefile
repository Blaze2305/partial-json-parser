.PHONY: build test wasm

BASE_URL ?= "local"

build: wasm
	mv wasm/main.wasm .
	bundle exec jekyll build  --baseurl ${BASE_URL}

wasm:
	make -C wasm build-wasm 



