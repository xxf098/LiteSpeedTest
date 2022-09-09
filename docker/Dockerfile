# sudo docker build   --network=host  -t lite:0.1 -f ./docker/Dockerfile ./
FROM golang:1.18 as build
ADD / /go/lite
WORKDIR /go/lite
RUN go env -w GOPROXY=https://goproxy.cn,direct
RUN go mod tidy
RUN make dockerbin

# FROM gcr.io/distroless/static-debian11
FROM gcr.lank8s.cn/distroless/static-debian11
COPY --from=build /go/lite/bin/lite-dockerbin /
EXPOSE 10888
CMD [ "/lite-dockerbin" ]