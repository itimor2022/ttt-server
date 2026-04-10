FROM golang:1.20 as build

ENV GOPROXY https://goproxy.cn,direct
ENV GO111MODULE on


WORKDIR /go/cache


ADD go.mod .
ADD go.sum .
RUN go mod download

WORKDIR /go/release



# RUN apt-get update && \
#       apt-get -y install ca-certificates 

ADD . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags='-w -extldflags "-static"' -installsuffix cgo -o app ./main.go


FROM alpine as prod
# Import the user and group files from the builder.
COPY --from=build /etc/passwd /etc/passwd
COPY --from=build /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
RUN \ 
    mkdir -p /usr/share/zoneinfo/Asia && \
    ln -s /etc/localtime /usr/share/zoneinfo/Asia/Shanghai
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
WORKDIR /home
COPY --from=build /go/release/app /home
COPY --from=build /go/release/assets /home/assets
COPY --from=build /go/release/configs /home/configs
RUN echo "Asia/Shanghai" > /etc/timezone
ENV TZ=Asia/Shanghai

# 不加  apk add ca-certificates  apns2推送将请求错误
# RUN  apk add ca-certificates 

ENTRYPOINT ["/home/app"]
