#!/bin/bash

# 컨테이너 빌드 및 백그라운드에서 실행
docker-compose up -d --build 

# 로그 출력 (타임스탬프 포함, 실시간 업데이트)
docker-compose logs -t -f