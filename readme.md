openssl genrsa -out private_key.pem 2048
openssl rsa -in private_key.pem -pubout -out public_key.pem




<!-- ## AWS 배포 가이드

이 문서는 프로젝트를 AWS 환경에 배포하기 위한 상세한 가이드를 제공합니다.

## 목차

1. [키보안]
1. [VPC 구성](#1-vpc-구성)
2. [보안 그룹 강화](#2-보안-그룹-강화)
3. [EC2 인스턴스 강화](#3-ec2-인스턴스-강화)
4. [IAM 역할 및 정책 설정](#4-iam-역할-및-정책-설정)
5. [로깅 및 모니터링 설정](#5-로깅-및-모니터링-설정)
6. [암호화 및 키 관리](#6-암호화-및-키-관리)
7. [네트워크 ACL 설정](#7-네트워크-acl-설정)
8. [배포 스크립트 보안 강화](#8-배포-스크립트-보안-강화)



## 1. 키보안

1. 키 생성 시 개별 secret값을 전달

키 생성 요청(key_gen)을 진행할 시 {secret값을 전달.}



## 1. VPC 구성

1. 새로운 VPC 생성
   - CIDR 블록: 10.0.0.0/16

2. 서브넷 생성
   - 퍼블릭 서브넷: 10.0.1.0/24, 10.0.2.0/24
   - 프라이빗 서브넷: 10.0.3.0/24, 10.0.4.0/24

3. 인터넷 게이트웨이 연결
   - VPC에 인터넷 게이트웨이 연결

4. NAT 게이트웨이 설정
   - 프라이빗 서브넷의 아웃바운드 인터넷 접속용

5. 라우팅 테이블 구성
   - 퍼블릭 서브넷: 인터넷 게이트웨이로 라우팅
   - 프라이빗 서브넷: NAT 게이트웨이로 라우팅

## 2. 보안 그룹 강화

1. 게이트웨이 보안 그룹
   ```
   인바운드:
   - HTTP (80): 0.0.0.0/0
   - HTTPS (443): 0.0.0.0/0
   - SSH (22): 관리자 IP만
   아웃바운드:
   - 모든 트래픽: 0.0.0.0/0
   ```

2. Alice/Bob 보안 그룹
   ```
   인바운드:
   - 커스텀 TCP: 게이트웨이 보안 그룹으로부터
   - SSH (22): 관리자 IP만
   아웃바운드:
   - 모든 트래픽: 0.0.0.0/0
   ```

## 3. EC2 인스턴스 강화

1. EC2 인스턴스 배치
   - 게이트웨이: 퍼블릭 서브넷
   - Alice/Bob: 프라이빗 서브넷

2. EC2 인스턴스 설정
   ```bash
   # 시스템 업데이트
   sudo yum update -y

   # 필요한 패키지 설치
   sudo yum install -y golang git

   # SSH 보안 강화
   sudo sed -i 's/#PermitRootLogin yes/PermitRootLogin no/' /etc/ssh/sshd_config
   sudo sed -i 's/PasswordAuthentication yes/PasswordAuthentication no/' /etc/ssh/sshd_config
   sudo systemctl restart sshd

   # 방화벽 설정
   sudo yum install -y iptables-services
   sudo systemctl enable iptables
   sudo systemctl start iptables

   # iptables 규칙 설정 (예시)
   sudo iptables -A INPUT -p tcp --dport 22 -j ACCEPT
   sudo iptables -A INPUT -p tcp --dport 80 -j ACCEPT
   sudo iptables -A INPUT -p tcp --dport 443 -j ACCEPT
   sudo iptables -A INPUT -j DROP
   sudo service iptables save
   ```

## 4. IAM 역할 및 정책 설정

1. EC2용 IAM 역할 생성
   - AmazonEC2RoleforSSM 정책 연결 (시스템 관리용)

2. 커스텀 정책 생성 및 연결
   ```json
   {
     "Version": "2012-10-17",
     "Statement": [
       {
         "Effect": "Allow",
         "Action": [
           "s3:GetObject",
           "s3:PutObject"
         ],
         "Resource": "arn:aws:s3:::your-bucket-name/*"
       },
       {
         "Effect": "Allow",
         "Action": [
           "kms:Encrypt",
           "kms:Decrypt"
         ],
         "Resource": "arn:aws:kms:region:account-id:key/key-id"
       }
     ]
   }
   ```

## 5. 로깅 및 모니터링 설정

1. CloudWatch 에이전트 설치
   ```bash
   sudo yum install -y amazon-cloudwatch-agent
   ```

2. CloudWatch 설정
   ```json
   {
     "agent": {
       "metrics_collection_interval": 60,
       "run_as_user": "root"
     },
     "logs": {
       "logs_collected": {
         "files": {
           "collect_list": [
             {
               "file_path": "/var/log/your-service.log",
               "log_group_name": "/ec2/your-service",
               "log_stream_name": "{instance_id}"
             }
           ]
         }
       }
     },
     "metrics": {
       "metrics_collected": {
         "mem": {
           "measurement": ["mem_used_percent"]
         },
         "swap": {
           "measurement": ["swap_used_percent"]
         }
       }
     }
   }
   ```

3. CloudWatch 에이전트 시작
   ```bash
   sudo /opt/aws/amazon-cloudwatch-agent/bin/amazon-cloudwatch-agent-ctl -a fetch-config -m ec2 -c file:/opt/aws/amazon-cloudwatch-agent/etc/amazon-cloudwatch-agent.json -s
   ```

## 6. 암호화 및 키 관리

1. AWS KMS 사용
   - 새로운 KMS 키 생성
   - IAM 역할에 KMS 키 사용 권한 부여

2. 설정 파일 암호화
   ```bash
   # 암호화
   aws kms encrypt --key-id alias/your-key-alias --plaintext fileb://config.json --output text --query CiphertextBlob | base64 --decode > config.json.encrypted

   # 복호화
   aws kms decrypt --ciphertext-blob fileb://config.json.encrypted --output text --query Plaintext | base64 --decode > config.json
   ```

## 7. 네트워크 ACL 설정

1. 인바운드 규칙
   ```
   100: HTTP(80)  허용 0.0.0.0/0
   110: HTTPS(443) 허용 0.0.0.0/0
   120: SSH(22)   허용 관리자 IP
   * : 모두 거부
   ```

2. 아웃바운드 규칙
   ```
   100: HTTP(80)  허용 0.0.0.0/0
   110: HTTPS(443) 허용 0.0.0.0/0
   120: Ephemeral Ports(1024-65535) 허용 0.0.0.0/0
   * : 모두 거부
   ```

## 8. 배포 스크립트 보안 강화

1. 비밀 정보 관리
   - AWS Secrets Manager 사용
   ```bash
   # 비밀 정보 저장
   aws secretsmanager create-secret --name "your-secret-name" --secret-string '{"key":"value"}'

   # 비밀 정보 조회
   aws secretsmanager get-secret-value --secret-id "your-secret-name" --query SecretString --output text
   ```

2. 배포 스크립트 수정
   ```bash
   #!/bin/bash

   SERVICE_NAME=$1
   INSTANCE_IP=$2

   # AWS Secrets Manager에서 SSH 키 가져오기
   SSH_KEY=$(aws secretsmanager get-secret-value --secret-id "ec2-ssh-key" --query SecretString --output text)
   echo "$SSH_KEY" > temp_ssh_key
   chmod 600 temp_ssh_key

   # 배포 작업
   scp -i temp_ssh_key -r ./cmd/$SERVICE_NAME ec2-user@$INSTANCE_IP:/home/ec2-user/
   ssh -i temp_ssh_key ec2-user@$INSTANCE_IP << EOF
       cd /home/ec2-user/$SERVICE_NAME
       go build -o main
       sudo systemctl stop $SERVICE_NAME || true
       sudo cp main /usr/local/bin/$SERVICE_NAME
       sudo systemctl start $SERVICE_NAME
   EOF

   # 임시 SSH 키 파일 삭제
   rm temp_ssh_key

   echo "Deployment of $SERVICE_NAME to $INSTANCE_IP completed."
   ```

이 가이드를 따라 설정을 완료하면, 안전하고 효율적인 AWS 환경에서 프로젝트를 운영할 수 있습니다. 각 단계를 신중히 적용하고, 필요에 따라 조정하세요. 또한, 정기적인 보안 감사와 업데이트를 통해 시스템의 보안을 지속적으로 유지하는 것이 중요합니다. -->