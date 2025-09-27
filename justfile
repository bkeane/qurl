# List help
[private]
default:
    @just --list --unsorted

# build and install qurl to ~/.local/bin
install:
    go build -o ~/.local/bin/qurl cmd/qurl/main.go

# test release build locally (snapshot)
snapshot:
    goreleaser release --snapshot --clean --skip=publish