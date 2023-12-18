# xconfadmin

This project is to implement a configuration management server. RDK devices download configurations from this server during bootup or notified when updates are available.

## Install go

This project is written and tested with Go **1.15**.

## Build the binary
```shell
cd .../xconfadmin
make
```
**bin/xconfadmin-linux-amd64** will be created. 

## Run the application
A configuration file can be passed as an argument when the application starts. config/sample_xconfwebconfig.conf is an example. 


```shell
mkdir -p /app/logs/xconfadmin
cd .../xconfadmin
bin/xconfadmin-linux-amd64 -f config/sample_xconfwebconfig.conf
```

```shell
curl http://localhost:9000/api/v1/version
{"status":200,"message":"OK","data":{"code_git_commit":"2ac7ff4","build_time":"Thu Feb 14 01:57:26 2019 UTC","binary_version":"317f2d4","binary_branch":"develop","binary_build_time":"2021-02-10_18:26:49_UTC"}}
```

## Run the tests
To run all of the tests in xconfadmin project:
```shell
cd .../xconfadmin
make test
```

