#!/bin/bash

# Copyright 2021 The CloudEvents Authors
# SPDX-License-Identifier: Apache-2.0

set -o errexit
set -o nounset

# Makes sure that everything is killed after the script is stopped
(trap 'kill 0' SIGINT; docker-compose up & go run producer/producer.go & go run intermediary/intermediary.go & go run consumer/consumer.go)