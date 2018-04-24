tests_dir := tests
tester := ${tests_dir}/tech-db-forum
report := ${tests_dir}/report.html

docker_name := docker_forum_tp
docker_tag := 1.0




build:
	./scripts/build.sh

app_run:
	./server.app


func-test:
	./${tester} func --wait=3 --keep -u http://localhost:5000/ -r tests/report.html

show-report:
	firefox tests/report.html https://tech-db-forum.bozaro.ru/ & echo "report and api-list"

clear:
	rm -rf vendor ${app}

docker-no-cache:
	docker build --no-cache -t ${docker_name}:${docker_tag} -f Dockerfile ./

docker:
	docker build -t ${docker_name}:${docker_tag} -f Dockerfile ./

run:
	docker run -p 5000:5000 --rm -d -it --name forum_tp ${docker_name}:${docker_tag}

stop:
	docker stop forum_tp

logs:
	docker logs forum_tp



delete-container:
	docker images
	docker rmi ${docker_name}:${docker_tag}
	docker images