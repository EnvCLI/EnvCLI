# Build & Test EnvCLI

## Download the Dependencies

```
envcli run dep ensure
```

## Build the Binaries (Windows/Linux/Mac)

```
envcli --env GOOS=windows --env GOARCH=386 run go build -o build/envcli_windows_386.exe src/*
envcli --env GOOS=windows --env GOARCH=amd64 run go build -o build/envcli_windows_amd64.exe src/*
envcli --env GOOS=linux --env GOARCH=386 run go build -o build/envcli_linux_386 src/*
envcli --env GOOS=linux --env GOARCH=amd64 run go build -o build/envcli_linux_amd64 src/*
envcli --env GOOS=darwin --env GOARCH=386 run go build -o build/envcli_darwin_386 src/*
envcli --env GOOS=darwin --env GOARCH=amd64 run go build -o build/envcli_darwin_amd64 src/*
```

## Run EnvCLI (to test your changes before building the final binaries)

Use this to test if EnvCLI is working as expected with increased logging.

```
envcli run go run src/* --loglevel=debug help
```
