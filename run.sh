#!/usr/bin/env bash

go run backend/main.go -p 7131 &
http-server -p 7132 ./
