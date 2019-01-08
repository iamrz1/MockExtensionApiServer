### Instructions:
```gotemplate
go run KubeApiServer/main.go
curl https://127.0.0.1:8443/core/Pod --cacert databaseserver-ca.crt --cert databaseserver-rezoan.crt --key databaseserver-rezoan.key
>> 2019/01/08 15:50:49 &{0xc0002d8200} Resource:  newReq


go run DatabaseServer/main.go
curl https://127.0.0.2:8443/db/newReq --cacert databaseserver-ca.crt --cert databaseserver-rezoan.crt --key databaseserver-rezoan.key
>> 2019/01/08 15:50:49 &{0xc0002d8200} Resource:  newReq

```