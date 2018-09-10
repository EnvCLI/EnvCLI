# Build & Test EnvCLI

## Download the Dependencies

```
envcli run dep ensure
```

## Build the Binaries (Windows/Linux/Mac)

```
envcli run --env GOOS=windows --env GOARCH=386 go build -o build/envcli_windows_386.exe src/*
envcli run --env GOOS=windows --env GOARCH=amd64 go build -o build/envcli_windows_amd64.exe src/*
envcli run --env GOOS=linux --env GOARCH=386 go build -o build/envcli_linux_386 src/*
envcli run --env GOOS=linux --env GOARCH=amd64 go build -o build/envcli_linux_amd64 src/*
envcli run --env GOOS=darwin --env GOARCH=386 go build -o build/envcli_darwin_386 src/*
envcli run --env GOOS=darwin --env GOARCH=amd64 go build -o build/envcli_darwin_amd64 src/*
```

## Run EnvCLI (to test your changes before building the final binaries)

Use this to test if EnvCLI is working as expected with increased logging.

```
envcli run go run src/* --loglevel=debug help
```
