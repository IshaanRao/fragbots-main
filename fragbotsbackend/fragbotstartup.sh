#!/bin/bash
amazon-linux-extras install docker -y
service docker start
dockerd -H unix:///var/run/docker.sock -H tcp://0.0.0.0:2375
docker network create fragnet
docker pull ishaanrao/fragbots:latest
