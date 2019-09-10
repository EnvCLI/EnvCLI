#!/usr/bin/env bash

# generate embedded binary data
echo "--> Generating embedded binary data"
go-bindata -o pkg/aliases/scripts.go -pkg aliases scripts/
