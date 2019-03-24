# Makefile used at remote server.

auth:
	aws ecr get-login --no-include-email | bash

stop:
        docker stop biblia2y || :
        docker rm biblia2y || :

pull:
        docker pull 518205865037.dkr.ecr.us-east-1.amazonaws.com/biblia2y:master_latest

restart: pull stop
        docker run --name biblia2y -d -p443:10101 --mount source=/home/ubuntu/db,target=/db,type=bind --mount source=/etc/letsencrypt/live/biblia2y.qradium.com/fullchain.pem,target=/etc/letsencrypt/live/biblia2y.qradium.com/fullchain.pem,type=bind,readonly --mount source=/etc/letsencrypt/live/biblia2y.qradium.com/privkey.pem,target=/etc/letsencrypt/live/biblia2y.qradium.com/privkey.pem,type=bind,readonly 518205865037.dkr.ecr.us-east-1.amazonaws.com/biblia2y:master_latest


renew:
        certbot renew --pre-hook "docker stop biblia2y && docker rm biblia2y" --post-hook "docker pull 518205865037.dkr.ecr.us-east-1.amazonaws.com/biblia2y:master_latest && docker run --name biblia2y -d -p443:10101 --mount source=/home/ubuntu/db,target=/db,type=bind --mount source=/etc/letsencrypt/live/biblia2y.qradium.com/fullchain.pem,target=/etc/letsencrypt/live/biblia2y.qradium.com/fullchain.pem,type=bind,readonly --mount source=/etc/letsencrypt/live/biblia2y.qradium.com/privkey.pem,target=/etc/letsencrypt/live/biblia2y.qradium.com/privkey.pem,type=bind,readonly 518205865037.dkr.ecr.us-east-1.amazonaws.com/biblia2y:master_latest"