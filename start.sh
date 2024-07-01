#!/bin/bash

if ! test -f ./a2b-api-go; then
	go build -o a2b-api-go .
fi
# redirect stdout & stderr to ./a2b-api.log (append mode)
./a2b-api-go >>./a2b-api.log 2>&1 &
