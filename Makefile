build:
	go build -o bin/memtest .

run:
	go run main.go

image-build:
	docker build -t registry.k8s.io/memtest:test .

setup-env:
	./env/setup_env.sh

teardown-env:
	./env/teardown_env.sh


