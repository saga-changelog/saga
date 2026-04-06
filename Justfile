set shell := ["bash", "-c"]

_default:
    just --list

fmt:
    go fmt ./...

setup:
    mise install
    prek install
    go mod download

# Run theme test across all color profiles
theme-test:
    go run . theme-test --profile truecolor
    go run . theme-test --profile ansi256
    go run . theme-test --profile ansi
    go run . theme-test --profile ascii

# Install saga and all couriers to $GOPATH/bin
install:
    go install .
    go install ./couriers/saga-courier-local-file
    go install ./couriers/saga-courier-local-stdout
    go install ./couriers/saga-courier-slack-legacy
    go install ./couriers/saga-courier-basecamp-messageboard
