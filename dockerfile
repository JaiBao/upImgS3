# 使用 Go 官方鏡像作為構建環境
FROM golang:1.20 as builder

# 設置工作目錄，之後的指令都會在這個目錄下執行
WORKDIR /app

# 將 go.mod 和 go.sum 文件複製到容器中
COPY go.mod go.sum ./

# 下載項目依賴。利用 Docker 緩存層，除非 go.mod 和 go.sum 改變，否則不需要重複下載依賴
RUN go mod download

# 將項目中的所有文件複製到容器中
COPY . .

# 構建應用程序
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o myapp .

# 使用 alpine 作為最終運行時鏡像
FROM alpine:latest  

# 安裝 ca-certificates
RUN apk --no-cache add ca-certificates

# 將工作目錄設置為 /root/
WORKDIR /root/

# 從構建階段的 /app 目錄中複製構建好的二進制文件和 .env 文件到當前目錄
COPY --from=builder /app/myapp .
COPY --from=builder /app/.env .

# 配置容器啟動時運行的命令
ENTRYPOINT ["./myapp"]

# 暴露端口
EXPOSE 8080
