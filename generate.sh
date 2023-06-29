#!/bin/zsh
protoc proto/chord.proto --go_out=. --go-grpc_out=.
protoc proto/aplication.proto --go_out=. --go-grpc_out=.