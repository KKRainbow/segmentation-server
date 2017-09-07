#!/bin/bash
sudo docker stop segserver
sudo docker rm segserver
sudo nvidia-docker run -it -v `pwd`:/go/src/github.com/KKRainbow/segmentation-server --env CUDA_VISIBLE_DEVICES=1 --env GOPATH=/go -v /mnt/cephfs:/cephfs -p 8888:8888 --name=segserver r.fds.so/segserver:latest /bin/bash
