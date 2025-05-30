# draw-hub
AI 绘图服务，对接了多家不同的供应商

## 主要功能
1. 接收绘图请求，生成任务，按序请求供应商接口，直到某一家成功

2. 查询绘图状态

3. 文件上传

模型:
✅ gpt-4o-image
✅ gpt-image-1

绘图供应商:
✅ [Geek](https://geekai.dev/)
✅ [Tuzi](https://api.v3.cm/)
✅ [V3](https://api.tu-zi.com/)

## 初衷
1. 寻找价格较低的生图供应商，但接口不稳定，想要加强服务稳定性
    
2. 对接多家供应商，支持更多模型

## 快速开始
### 前置条件
- 一个或多个绘图供应商服务（Geek、Tuzi、V3）
- MySQL 5.7及以上
- Golang 1.23及以上
   
```shell
git clone https://github.com/reusedev/draw-hub.git
cd draw-hub
# 配置MySQL、绘图供应商及请求优先级
vim config.yml
# 下载依赖
go mod download
# 运行
go run main.go
```
### Docker 运行
```shell
docker build -t draw-hub .
docker run -d -p 80:80 -v /path/to/config.yml:/app/config.yml draw-hub
```
