set shell := ["bash", "-c"]

_default:
    just --list

fmt:
    go fmt ./...

setup:
    mise install
    prek install
    go mod download

# Install saga and all couriers to $GOPATH/bin
install:
    go install .
    go install ./couriers/saga-courier-stdout
    go install ./couriers/saga-courier-slack
    go install ./couriers/saga-courier-basecamp
