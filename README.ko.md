# AMO blockchain을 위한 Tendermint ABCI 앱
This document is available in [English](README.md) also.

## 개요
현재 AMO 블록체인은 구현상
[Tendermint](https://github.com/tendermint/tendermint)을 기본 합의 구조로
사용한다. Tendermint는 블록체인 노드들 간의 P2P 통신과, validator 노드들 사이의
합의 알고리즘 진행, 클라이언트로부터의 요청을 처리하는 RPC 서버 구동을
담당한다. 그러나, Tendermint만으로 온전한 블록체인으로 동작하지는 않으며 각
블록내 정보, 즉 포함된 거래들을 해석하는 ABCI 앱이 필요하다. ABCI 앱은 거래, 즉
상태 변화와, 블록체인의 상태, validator 제어, 그리고 블록체인의 상태에 대한
클라이언트로부터의 조회 요청을 처리한다. 이 repository는 *AMO 블록체인을 위한
Tendermint ABCI 앱*을 구현하는 코드(`amoabci`)와 그외 필요한 스크립트의
조합으로 구성된다.

## 컴파일된 바이너리 설치
TBA

## 소스코드로부터 설치

### 사전 조건
소스코드로부터 컴파일을 하기 위해서는 다음을 설치해야 한다:
* [git](https://git-scm.com)
* [make](https://www.gnu.org/software/make/)
  * Debian이나 Ubuntu 리눅스의 경우에는 `build-essential` 패키지를 설치한다.
  * MacOS의 경우에는 Xcode의 `make`를 사용허가나 [Homebrew](https://brew.sh)를
	통해 GNU Make를 설치할 수 있다.
* [golang](https://golang.org/dl/)
  * 경우에 따라서 `GOPATH`와 `GOBIN` 환경변수를 수동으로 설정해 줘야 할 수
	있다. 이후 더 진행하기 전에 이 변수들을 확인하도록 한다.
* [leveldb](https://github.com/google/leveldb)
  * Debian이나 Ubuntu 리눅수의 경우에는 `libleveldb-dev` 패키지를 설치한다.
  * 컴파일하는 서버와 실행하는 서버가 다를 경우 실행하는 서버에는
	`libleveldb1v5` 패키지를 설치한다.

데몬 프로그램들을 docker 컨테이너에서 실행하거나 docker를 필요로 하는
테스트들을 실행하기 위해서는 다음을 설치한다:
* [docker](https://www.docker.com) (Debian이나 Ubuntu에서는 `docker.io`를 설치)
* [docker-compose](https://www.docker.com)

### tendermint 설치
AMO 블록체인을 위한 ABCI 앱을 실행하려면 같은 호스트에서
[tendermint](https://github.com/tendermint/tendermint) 데몬이 실행되고 있어야
한다. 따라서 tendermint를 먼저 설치하도록 한다. 현재 버전의 AMO ABCI 앱은
tendermint v0.32.3이 필요하다.

다음 명령을 실행해서 tendermint 데몬을 설치한다:
```bash
mkdir -p $GOPATH/src/github.com/tendermint
cd $GOPATH/src/github.com/tendermint
git clone https://github.com/tendermint/tendermint
cd tendermint
make get_tools
make install_c
```

### amod 설치
다음 명령을 실행해서 amod를 설치한다:
```bash
mkdir -p $GOPATH/src/github.com/amolabs
cd $GOPATH/src/github.com/amolabs
git clone https://github.com/amolabs/amoabci
cd amoabci
make get_tools
make install
```

### amocli 설치
`amocli`가 없어도 필요한 데몬들을 실행할 수 있지만, 현재 일어나고 있는 상황을
확인하이 위해 블록체인 노드 데몬들의 상태를 확인해야 할 수 있다. AMO Labs에서는
참조 구현의 일환으로 AMO client(`amocli`)를 제공하며 이 프로그램을 설치해서 AMO
블록체인 노드들과 통신을 할 수 있다.
```bash
mkdir -p $GOPATH/src/github.com/amolabs
cd $GOPATH/src/github.com/amolabs
git clone https://github.com/amolabs/amo-client-go
cd amo-client-go
make install
```

더 자세한 사항은 [amo-client-go](https://github.com/amolabs/amo-client-go)를
참조하도록 한다.

## 실행 준비
### 네트워크 정보 수집
AMO 블록체인 노드는 네트워크 응용프로그램이다. 다른 노드들에 연결되지 않으면
별다은 의미 있는 동작을 하지 못한다. 따라서 첫번째로 알아내야 하는 것은 AMO
블록체인 네트워크에 있는 다른 노드들의 주소이다. 네트워크에 있는 여러 노드
중에서 **seed** 노드라 불리는 노드들 중 하나에 연결하는 것을 권장한다. 만약
적당한 seed 노드가 없다면 **이웃**을 충분히 확보한 노드에 연결한다.

* 메인넷 정보: http://mainnet.amolabs.io
* 테스트넷 정보: http://testnet.amolabs.io

*로컬에서 테스트넷을 실행하기 위한 정보는 다음에서 확인한다: TBA*

### genesis.json 확보
블록체인은 끊임 없이 변화하는 [상태
기계](https://en.wikipedia.org/wiki/Finite-state_machine)이다. 따라서
블록체인의 초기 상태가 무엇인지 알아내야 한다. AMO 블록체인은 tendermint의
방식을 사용하므로 체인의 초기 상태를 정의하는 `genesis.json` 파일을 확보해야
한다.

* 메인넷 정보: http://mainnet.amolabs.io
* 테스트넷 정보: http://testnet.amolabs.io

### 데이터 디렉토리 준비
`tendermint`와 `amod` 모두 데이터 디렉토리가 필요하며, 여기에 설정 파일과 내부
데이터베이스를 저장한다. `tendermint`와 `amod`가 데이터 디렉토리를 공유하지는
않지만, 둘의 디렉토리를 합한 것이 결국 AMO 블록체인의 완전한 스냅샷이 된다.
따라서 디렉토리 구조는 다음과 같은 형태가 되도록 할 것을 권장한다:
```
(node_data_root)
├── amo
│   └── data
└── tendermint
    ├── config
    └── data
```

여기에서 `dataroot/tendermint/config` 디렉토리에는 특히 `node_key.json`과
`priv_validator_key.json`과 같은 민감한 파일들이 저장한다. 이 파일들은 읽기
권한을 조정하여 안전하게 저장해야 한다. **이는 docker 컨테이너로 데몬들을
실행하는 경우에도 해당한다.**

### 필요한 파일들 준비
`tendermint`는 그 동작을 위해 몇가지 파일들을 필요로 한다:
- `config.toml`<sup>&dagger;</sup>: 설정
- `genesis.json`<sup>&dagger;</sup>: 블록체인과 앱의 초기 상태
- `node_key.json`<sup>&dagger;&dagger;</sup>: p2p 연결을 위한 노드 키
- `priv_validator_key.json`<sup>&dagger;&dagger;</sup>: 합의 과정을 위한
  validator 키

&dagger; 이 파일들은 `tendermint`나 `amod`를 실행하기 전에 먼저 준비해야 한다.
주목해야 할 몇가지 설정 옵션들은 다음과 같다:
- `moniker`
- `rpc.laddr`
- `rpc.cors_allowed_origins`
- `p2p.laddr`
- `p2p.external_adderess`
- `p2p.seeds`
- `p2p.persistent_peers`

보다 자세한 정보는 [Tendermint
문서](https://tendermint.com/docs/tendermint-core/configuration.html)를
참조한다.

&dagger;&dagger; 이 파일들은 미리 준비돼지 않았다면 `tendermint`가 스스로
생성할 수 있다. 다만, 특정한 키를 사용하고자 한다면 실행 전에 미리 준비해야
한다. 가능한 방법중 한가지는 `tendermint init` 명령으로 키들을 생성한 후
`config.toml`과 `genesis.json` 파일이 있는 설정 디렉토리에 넣어 두는 것이다.

## 데몬 직접 실행
`amod`를 먼저 실행하는 것이 안전하다.
```bash
amod --home <dataroot>/amo run
```
데몬을 백그라운드 모드로 실행하려면 `amod run &`와 같이 한다. 여기에서
`<dataroot>`는 앞서 준비한 데이터 디렉토리이다. `amod`는 유입되는 ABCI 연결을
위해 포트 26658을 연다.

그 이후 `tendermint`를 실행한다.
```bash
tendermint --home <dataroot>/tendermint node
```
물론 이번에도 `tendermint node &`와 같이 하여 데몬을 백그라운드 모드로 실행할
수 있다. 여기에서 `<dataroot>`는 앞서 준비한 데이터 디렉토리이다.
`tendermint`는 유입되는 P2P 연결을 위해 포트 26656을 열고, 유입되는 RPC 연결을
위해 포트 26657을 연다. 이 데몬은 ABCI 데몬, 우리의 경우 `amod`에 연결하기
위해서 localhost의 포트 26658로 연결을 시도한다.

## docker로 데몬 실행
### 사전 조건
* [docker](https://www.docker.com) (Debian이나 Ubuntu에서는 `docker.io`를 설치)

### docker 이미지 생성
AMO Labs에서 배포하는 `amod`의 공식 docker 이미지(`amolabs/amod`)는 [Docker
hub](https://hub.docker.com)에서 다운로드할 수 있다. 물론 로컬 docker 이미지를
직접 생성할 수도 있다.

`amod`의 docker 이미지를 생성하기 전에 `tendermint`를 먼저 컴파일해야 한다.
다음과 같이 하여 직접 할 수도 있고 `amod`의 Makefile을 통해 할 수도 있다. 직접
하는 경우는 다음과 같이 한다:
```bash
mkdir -p $GOPATH/src/github.com/tendermint
cd $GOPATH/src/github.com/tendermint
git clone https://github.com/tendermint/tendermint
cd tendermint
make get_tools
make build-linux
cp tendermint $GOPATH/src/github.com/amolabs/amoabci/
```

`amod`의 docker 이미지를 생성하기 위해서는 다음과 같이 한다:
```bash
mkdir -p $GOPATH/src/github.com/amolabs
cd $GOPATH/src/github.com/amolabs
git clone https://github.com/amolabs/amoabci
cd amoabci
make get_tools
make docker
```
이미지는 `amolabs/amod:latest`로 태그된다. 이 이미지는 `tendermint`와 `amod`를
모두 포함하고 있기 때문에 하나의 이미지(따라서 하나의 컨테이너)만 있으면 된다.

### 실행
컨테이너에서 데몬들을 실행하기 위해서는 다음과 같이 한다:
```bash
docker run -it --rm -p 26656-26657 -v <dataroot>/tendermint:/tendermint:Z -v <dataroot>/amo:/amo:Z -d amolabs/amod:latest
```
위에 사용된 명령행 옵션들은 다음과 같은 의미를 갖는다:
- `-it`: 터미널 연결 확보
- `--rm`: 데몬들이 중지된 후에 컨테이너 삭제
- `-p 26656-26657`: 컨테이너의 포트를 호스트 머신에 연결. 이를 통해 네트워크의
  다른 노드들이 이 노드에 정상적으로 연결할 수 있게 된다.
- `-v <dataroot>/tendermint:/tendermint:Z`: tendermint 데이터 디렉토리 연결
  **`<dataroot>` 는 절대 경로여야 한다.**
- `-v <dataroot>/amo:/amo:Z`: amod 데이터 디렉토리 연결
  **`<dataroot>` 는 절대 경로여야 한다.**
- `amolabs/amod:latest`: 컨테이너를 생성할 때 이 이미지를 사용

데몬들이 초기화하고 실행되는 동안 로그들이 정상적으로 표시되는지 확인한다.
