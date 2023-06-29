#!/bin/zsh
#Por si hay algun error a la hora de crear los pb.go es porque no tiene configurado el path de go-gprc
export PATH=$PATH:$(go env GOPATH)/bin
source ~/.zshrc