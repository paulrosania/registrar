all: toolchain

clean:
	rm -rf util/_toolchain/go

toolchain: util/_toolchain/go/bin/go

util/_toolchain/go/bin/go: util/_toolchain/bin/gonative
	cd util/_toolchain && rm -rf go && bin/gonative build -version=1.4.3 && ls | grep -v "^\(bin\|go\)$$" | xargs rm -r


util/_toolchain/bin/gonative:
	go build -o util/_toolchain/bin/gonative github.com/inconshreveable/gonative
