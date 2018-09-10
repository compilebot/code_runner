if [ "$1" == "nocontainer" ]; then
  echo "Starting new container.."
else
  echo "Restarting container..!"
  docker stop code_runner && docker rm code_runner
fi
docker build -t code_runner . && docker run --network br0 -d --name code_runner --env-file .env -v /var/run/docker.sock:/var/run/docker.sock -e DOCKER_API_VERSION=1.38 code_runner
