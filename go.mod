module github.com/playnb/mustang

go 1.12

require (
	github.com/cihub/seelog v0.0.0-20170130134532-f561c5e57575 // indirect
	github.com/gin-gonic/gin v1.4.0
	github.com/go-redis/redis v6.15.2+incompatible
	github.com/json-iterator/go v1.1.6
	github.com/spf13/viper v1.4.0
	gopkg.in/mgo.v2 v2.0.0-20180705113604-9856a29383ce
)

replace github.com/playnb/mustang => ../mustang
