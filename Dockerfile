# FROM kneerunjun/gogingonic:latest
FROM golang:1.15.11-alpine3.13
# from the vanilla image of go gin with mgo driver
# mapping for log files
ARG SRC
ARG LOG
ARG RUN
ARG ETC 
ARG BIN
# making all the specific directories, refer to the env file which has the values 
RUN mkdir -p ${SRC} && mkdir -p ${LOG} && mkdir -p ${RUN} && mkdir -p ${ETC} && mkdir -p /var/www/luminapp/pages
WORKDIR ${SRC}
COPY go.sum go.mod ./
RUN go mod download 
COPY . .
RUN go build -o ${BIN}/luminapi .
