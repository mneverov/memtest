build:
	GOEXPERIMENT=greenteagc go1.25rc1 build -o bin/memtest-gt .

buildnogreantea:
	GOEXPERIMENT=nogreenteagc go1.25rc1 build -o bin/memtest-old .

run:
	go run main.go

build-image:
	docker build -t registry.k8s.io/memtest:test .

setup-env:
	./env/setup_env.sh

teardown-env:
	./env/teardown_env.sh
