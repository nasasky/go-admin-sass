# nasa-go-admin


### 使用 gin gorm mysql jwt session

## 本地运行

#### 运行数据库服务
docker 方式
```shell
docker compose -f docker-compose-env.yaml 
```
或者自行修改 .env 文件 Mysql 连接参数，并导入 init.sql 数据库及表结构

#####  运行前端
```shell
cd vue-naive-front && npm install && npm run dev
```
##### 运行后端
```shell
go run main.go
```
其他

要更新所有的 Go 模块，可以使用以下命令：

go get -u ./...

接着，运行以下命令来整理 go.mod 文件并下载更新后的依赖项：

go mod tidy





D:\mycode\nasa-go-admin>set CGO_ENABLED=0

D:\mycode\nasa-go-admin>go env -w GOOS=linux   //go env -w GOOS=windows

D:\mycode\nasa-go-admin>go env -w GOARCH=amd64

D:\mycode\nasa-go-admin>go build -o leishi-linux main.go





恢复本地开发环境的 GOOS 和 GOARCH 设置
在本地开发时，确保 GOOS 和 GOARCH 设置为 macOS 的默认值：


go env -w GOOS=darwin
go env -w GOARCH=amd64
重新运行 air
确保 air 使用的是针对 macOS 平台编译的二进制文件：


air
区分本地开发与交叉编译


本地开发：使用默认的 GOOS=darwin 和 GOARCH=amd64。


交叉编译：当需要为 Linux 平台生成二进制文件时，使用以下命令：


go env -w GOOS=linux
go env -w GOARCH=amd64
go build -o leishi-linux main.go
生成的 leishi-linux 文件可以在 Linux 环境中运行。


清理 air 的临时文件
如果 air 仍然尝试运行错误的二进制文件，清理 tmp 目录下的文件：


rm -rf tmp/*
然后重新运行 air。


