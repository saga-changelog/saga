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
    go install ./couriers/saga-courier-slack-app
    go install ./couriers/saga-courier-slack-legacy
    go install ./couriers/saga-courier-basecamp-messageboard

# Tag a release version for the root module and all courier modules
release version:
    #!/usr/bin/env bash
    set -euo pipefail
    git tag -a "{{version}}" -m "{{version}}"
    for dir in couriers/saga-courier-*/; do
        name="${dir%/}"
        git tag -a "${name}/{{version}}" -m "{{version}}"
    done
    echo "Tagged {{version}} for root and all couriers. Push with:"
    echo "  git push github {{version}} 'couriers/*/{{version}}'"
