
# go get -u github.com/swaggo/gin-swagger
# go get -u github.com/swaggo/files
# go get -u github.com/swaggo/swag/cmd/swag
# 在 pkg/plugins/modules/swagger目录下执行本脚本
cd ../../../../
swag init -g main.go  --exclude internal,pkg/comm/,pkg/service -o pkg/plugins/modules/swagger


# 删除docs.go 中的最后三行
sed -i '' '$d' pkg/plugins/modules/swagger/docs.go
sed -i '' '$d' pkg/plugins/modules/swagger/docs.go
sed -i '' '$d' pkg/plugins/modules/swagger/docs.go

# 向docs.go添加RegisterSwagger函数
echo "func RegisterSwagger() {
	swag.Register(swag.Name, &s{})
}" >> pkg/plugins/modules/swagger/docs.go

rm -rf pkg/plugins/modules/swagger/swagger.json
rm -rf pkg/plugins/modules/swagger/swagger.yaml