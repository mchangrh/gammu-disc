FROM golang:alpine as builder
WORKDIR /app
COPY . .
RUN apk add --no-cache upx
RUN go build
RUN mkdir /final
RUN upx --lzma --best -fq -o /final/gammu-disc /app/gammu-disc

FROM scratch
COPY --from=builder /final/gammu-disc /gammu-disc
ENTRYPOINT ["/gammu-disc"]