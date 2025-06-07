pipeline {
    agent any
    
    tools {
        go 'go1.19' // 使用 Jenkins 配置的 Go 工具
    }
    
    environment {
        GOOS = 'linux'
        GOARCH = 'amd64'
        GO111MODULE = 'on'
        APP_NAME = 'nasa-go-admin'
        REMOTE_DIR = '/home/apps/nasa-go-admin'
        SERVER_IP = '192.168.1.100'
    }
    
    stages {
        stage('Checkout') {
            steps {
                // 拉取代码
                checkout scm
            }
        }
        
        stage('Build') {
            steps {
                // 下载依赖并编译
                sh 'go mod tidy'
                sh 'go build -o ${APP_NAME}'
                sh 'chmod +x ${APP_NAME}'
            }
        }
        
        stage('Test') {
            steps {
                // 运行测试
                sh 'go test ./... -v'
            }
        }
        
        stage('Deploy') {
            steps {
                // 创建部署目录
                sh "ssh user@${SERVER_IP} 'mkdir -p ${REMOTE_DIR}'"
                
                // 传输应用程序和配置文件
                sh "scp ${APP_NAME} user@${SERVER_IP}:${REMOTE_DIR}/"
                sh "scp config.yaml user@${SERVER_IP}:${REMOTE_DIR}/"
                
                // 停止旧服务并启动新服务
                sh """
                    ssh user@${SERVER_IP} 'cd ${REMOTE_DIR} && ./stop.sh'
                    ssh user@${SERVER_IP} 'cd ${REMOTE_DIR} && ./start.sh'
                """
                
                // 检查服务是否正常启动
                sh "ssh user@${SERVER_IP} 'pgrep -f ${APP_NAME} || exit 1'"
            }
        }
    }
    
    post {
        success {
            echo '部署成功!'
        }
        failure {
            echo '部署失败!'
        }
    }
}