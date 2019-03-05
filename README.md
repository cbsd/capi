CBSD RESTFull API sample in golang


Init:

set GOPATH

go get

go run ./capi.go [ -l listen]


Endpoints:

*list bhyve domain*:

curl [-s] [-i] http://127.0.0.1:8080/api/v1/blist


*start (f11a) bhyve domain*:

curl -i -X POST http://127.0.0.1:8080/api/v1/bstart/f111a


*stop (f11a) bhyve domain*:

curl -i -X POST http://127.0.0.1:8080/api/v1/bstop/f111a


*remove (f11a) bhyve domain*:

curl -i -X POST http://127.0.0.1:8080/api/v1/bremove/f111a


*create new (f11a) bhyve domain (see *.json files for sample)*:

curl -X POST -H "Content-Type: application/json" -d @bhyve_create_minimal.json http://127.0.0.1:8080/api/v1/bcreate/f111a


This is a just simple example. Contributing is welcome!
