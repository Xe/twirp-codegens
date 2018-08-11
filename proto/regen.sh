#!/bin/sh

protoc --proto_path=$GOPATH/src:. \
       --twirp_out=. \
       --twirp_ln_out=. \
       --twirp_metrics_out=. \
       --go_out=. \
       ./test.proto
