main := src/main/main.go
app  := server.app


dependencies:
	dep ensure -update

vendor: dependencies
	dep ensure

easy_build:
	go build -o ${app} ${main}

build: vendor easy_build

${app}: easy_build

start: ${app}
	./${app}

start_in_docker: ${app}
	./${app} postgres://docker:docker@localhost:5432/forum_tp

test:
	./tests/tech-db-forum func -k -u http://localhost:5000/ -r tests/report.html

show:
	firefox tests/report.html

clear:
	rm -rf vendor ${app}

docker:
	docker build --no-cache -t forum-tp -f Dockerfile ./

run: forum-tp
	docker run -p 5000:5000 -rm -it --name forum-tp

forum-tp:
	docker build -t forum-tp -f Dockerfile ./
