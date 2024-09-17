# nasa-go-admin


### 使用 gin gorm mysql jwt session

## 本地运行

#### 运行数据库服务
docker 方式
```shell
docker compose -f docker-compose-env.yaml 
```
或者自行修改 .env 文件 Mysql 连接参数

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



#项目正在开发迭代中