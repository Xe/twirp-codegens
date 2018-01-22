#!/bin/sh

protoc --proto_path=$GOPATH/src:. \
       --twirp_out=. \
       --twirp_jsbrowser_out=. \
       --twirp_eclier_out=. \
       --go_out=. \
       ./test.proto
