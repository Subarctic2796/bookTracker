#!/usr/bin/env bash

build() {
    go build -o bookTracker src/*.go
}

run() {
    go run src/*.go $*
}

case "$1" in
    build) build ;;
    run)
        shift
        run $*
        ;;
    help) echo "Usage: ./build.sh [run|build|help]" ;;
    *)
        echo "Usage: ./build.sh [run|build|help]"
        exit 1
        ;;
esac
