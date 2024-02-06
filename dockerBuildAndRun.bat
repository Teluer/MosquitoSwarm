docker build -t mosquito-swarm:latest .
docker run -p 8008:8008 --restart unless-stopped mosquito-swarm