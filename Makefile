all:
	GOOS=js GOARCH=wasm go build -o ./public/ago.wasm main.go

cp_js:
	cp /usr/local/go/misc/wasm/wasm_exec.js ./public

start_server:
	./server/server &

publish:
	git checkout gh-pages
	git rebase master
	git push
	git checkout master
