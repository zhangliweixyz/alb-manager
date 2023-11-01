FROM harbor.weizhipin.com/bossv/golang:1.17.13
RUN unlink /etc/localtime
RUN ln -s /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
WORKDIR /root/
COPY . /root/
RUN wget -O "/root/conf/config.yaml" "https://consul.weizhipin.com/v1/kv/alb-manager/config-prod"
ENV GOPROXY=https://goproxy.cn,direct
RUN go mod tidy
RUN go build
RUN chmod +x alb-manager
CMD ["./alb-manager"]
