# xconfadmin

This project is to implement a configuration management server. RDK devices download configurations from this server during bootup or notified when updates are available.


## Install go

This project is written and tested with Go **1.23***

## Build the binary
```shell
cd $HOME/go/src/github.com/comcast-cl/xconfadmin
make
```
**bin/xconfadmin-linux-amd64** will be created. 

## Run the application
The application includes an API to notify RDK devices to download updated configurations from this server. A JWT token is required to communicate with service. The credentials are passed to the application through environment variables. A configuration file can be passed as an argument when the application starts. config/sample_xconfadmin.conf is an example. 


```shell
export SAT_CLIENT_ID='xxxxxx'
export SAT_CLIENT_SECRET='yyyyyy'
export SECURITY_TOKEN_KEY='zzzzzz'
mkdir -p /app/logs/xconfadmin
cd $HOME/go/src/github.com/comcast-cl/xconfadmin
bin/xconfadmin-linux-amd64 -f config/sample_xconfadmin.conf
```

```shell
curl http://localhost:9000/api/v1/version
{"status":200,"message":"OK","data":{"code_git_commit":"2ac7ff4","build_time":"Thu Feb 14 01:57:26 2019 UTC","binary_version":"317f2d4","binary_branch":"develop","binary_build_time":"2021-02-10_18:26:49_UTC"}}
```
