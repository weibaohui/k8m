
# go get -u github.com/swaggo/gin-swagger
# go get -u github.com/swaggo/files
# go get -u github.com/swaggo/swag/cmd/swag
cd ../
swag init -g main.go  --exclude internal,pkg/comm/,pkg/service -o swagger