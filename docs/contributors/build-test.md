# Build & Test EnvCLI

## Update all direct and indirect dependencies

```bash
go get -v all
```

## Embed alias scripts using GoBindata

```bash
go-bindata -o pkg/aliases/scripts.go -pkg aliases scripts/
```

## Build the Binaries (Windows/Linux/Mac)

```bash
envcli run --env GOOS=windows --env GOARCH=386 --env CGO_ENABLED=0 go build -o build/envcli_windows_386 -ldflags="-w" src/*
envcli run --env GOOS=windows --env GOARCH=amd64 --env CGO_ENABLED=0 go build -o build/envcli_windows_amd64 -ldflags="-w" src/*
envcli run --env GOOS=linux --env GOARCH=386 --env CGO_ENABLED=0 go build -o build/envcli_linux_386 -ldflags="-w" src/*
envcli run --env GOOS=linux --env GOARCH=amd64 --env CGO_ENABLED=0 go build -o build/envcli_linux_amd64 -ldflags="-w" src/*
envcli run --env GOOS=darwin --env GOARCH=386 --env CGO_ENABLED=0 go build -o build/envcli_darwin_386 -ldflags="-w" src/*
envcli run --env GOOS=darwin --env GOARCH=amd64 --env CGO_ENABLED=0 go build -o build/envcli_darwin_amd64 -ldflags="-w" src/*
```

## Build Binary on Windows for Local Testing

```bash
envcli run --env GOOS=windows --env GOARCH=amd64 --env CGO_ENABLED=0 go build -o build/envcli.exe -ldflags="-w" src/*
```

## Run EnvCLI (to test your changes before building the final binaries)

Use this to test if EnvCLI is working as expected with increased logging.

```bash
envcli run go run src/* --loglevel=debug help
```
