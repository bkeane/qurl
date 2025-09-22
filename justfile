# List help
[private]
default:
    @just --list --unsorted

# build and install monad to ~/.local/bin
install:
    go build -o ~/.local/bin/qurl cmd/qurl/main.go