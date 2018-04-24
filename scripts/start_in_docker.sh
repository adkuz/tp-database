#!/bin/bash

main=src/main/main.go
app=server.app

dir=$(pwd)

[[ $1 = "clear" ]] && rm -rf vendor ${app}

echo $dir

if [[ ! -d 'vendor' ]]; then
    dep ensure -update
    dep ensure
fi

[[ ! -f ${app} ]] && go build -o ${app} ${main}

service postgresql start

./${app} postgres://docker:docker@localhost:5432/forum_tp
