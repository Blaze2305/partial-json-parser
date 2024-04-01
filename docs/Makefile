.PHONY: build test wasm

wasm:
	make -C wasm build-wasm 

build: wasm
	mv wasm/main.wasm .
	bundle exec jekyll build 


