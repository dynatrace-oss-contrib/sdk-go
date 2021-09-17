#!/usr/bin/env bash

# Copyright 2021 The CloudEvents Authors
# SPDX-License-Identifier: Apache-2.0

set -o errexit
set -o nounset
set -o pipefail

# test/observability only
pushd ./test/observability

go test --tags=observability -v -timeout 15s

# Remove test only deps.
go mod tidy
popd