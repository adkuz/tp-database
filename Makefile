tests_dir := tests
tester := ${tests_dir}/tech-db-forum
report := ${tests_dir}/report.html

docker_name := docker_forum_tp
docker_tag := 1.0
container_name := forum_tp



build:
	./scripts/build.sh

app_run:
	./server.app


func-test:
	./${tester} func --wait=30 --keep -u http://localhost:5000/api/ -r tests/report.html

func-test-no-k:
	./${tester} func --wait=50 -u http://localhost:5000/api/ -r tests/report.html

fill-test:
	./tests/tech-db-forum fill --timeout=900

perform-test:
	./tests/tech-db-forum perf --validate=0.2 --duration=60


test_all: func-test fill-test perform-test



show-report:
	firefox file://$(shell pwd)/tests/report.html https://tech-db-forum.bozaro.ru/ & echo "report and api-list"

clear:
	rm -rf vendor ${app}

docker-no-cache:
	docker build --no-cache -t ${docker_name}:${docker_tag} -f Dockerfile ./

docker:
	docker build -t ${docker_name}:${docker_tag} -f Dockerfile ./

run:
	docker run -p 5000:5000 --rm -d -it --name ${container_name} ${docker_name}:${docker_tag}

run-no-d:
	docker run -p 5000:5000 --rm -it --name ${container_name} ${docker_name}:${docker_tag}

Nolan:
	docker exec -it ${container_name} bash


stop:
	docker stop ${container_name}

logs:
	docker logs ${container_name}


delete-container:
	docker images
	docker rmi ${docker_name}:${docker_tag}
	docker images