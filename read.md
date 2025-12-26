# Build & Run

## Binary

```sh
make build
make buildnogreentea
```

```sh
./bin/memtest-gt --record-iter=1000 --auto-stop --fr-out=snapshot-gt.trace
./bin/memtest-old --record-iter=1000 --auto-stop --fr-out=snapshot-nogt.trace
```

```sh
go tool trace snapshot-gt.trace
go tool trace snapshot-nogt.trace
```

## Kind Cluster

```sh
make setup-env
```
