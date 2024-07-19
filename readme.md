# TECDSA

Secure Two-party Threshold ECDSA from ECDSA Assumptions (https://github.com/coinbase/kryptology)

## 기능

- **클라이언트 등록**: 새로운 클라이언트를 보안 관리에 등록합니다.
- **키 생성**: 클라이언트를 위한 암호화 키를 생성합니다.
- **트랜잭션 서명**: 트랜잭션을 안전하게 서명합니다.
- **네트워크 관리**: 사용 가능한 네트워크 정보를 조회합니다.
- **문서 제공**: API 문서를 제공합니다.

## 시작하기

### 사전 요구사항

    - Go (버전 1.16 이상)


### 설치

- 저장소 클론:
```bash
git clone https://github.com/yourusername/tecdsa.git
```

- 스크립트 실행:
```bash
$ chmod +x start.sh
$ ./start.sh
```

### 도큐먼트

```
http://localhost:8080/docs
```


### 서버 접속 경로 
API 엔드포인트
| 메서드 | 엔드포인트            | 설명                                        |
|--------|----------------------|---------------------------------------------|
| POST   | `/register`          | 클라이언트의 보안 관리 등록(미완성)   |
| POST   | `/key_gen`           | 신규 주소 발급                |
| POST   | `/sign`              | 트랜잭션을 서명                       |
| GET    | `/networks`          | 사용 가능한 네트워크 목록을 조회합니다.        |
| GET    | `/docs/`             | API 문서를 제공합니다.                       |


