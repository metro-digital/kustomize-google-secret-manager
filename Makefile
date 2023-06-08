plugin = KGCPSecret
plugin_path = ${XDG_CONFIG_HOME}/kustomize/plugin/metro.digital/v1/kgcpsecret

all: clean build lint test

clean:
	-rm -f ${plugin}
	-find journey-test -name output.yaml -exec rm {} \;

build: clean
	go mod tidy
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o=${plugin} ./main/

install: build
	mkdir -p ${plugin_path}
	mv ${plugin} ${plugin_path}
	chmod +x ${plugin_path}/${plugin}

test:
	ginkgo -tags unitTests -r .

lint:
	golangci-lint run -c .golangci.yml ./...
