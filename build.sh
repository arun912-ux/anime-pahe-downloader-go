#!/bin/bash

set -e

echo "Building for all platforms"

for os in darwin linux windows; do
    for arch in arm64 amd64; do
        echo "Building for $os-$arch"
        if [ "$os" = "windows" ]; then
            GOOS=$os GOARCH=$arch go build -o build/animepahe-downloader-$os-$arch.exe .
        else
            GOOS=$os GOARCH=$arch go build -o build/animepahe-downloader-$os-$arch .
        fi
    done
done

