# FROM kneerunjun/gogingonic:latest
FROM golang:1.15.11-alpine3.13
# from the vanilla image of go gin with mgo driver
# mapping for log files
RUN mkdir -p /var/local/eensymachines/{logs,sockets,configs}
RUN mkdir -p $HOME/repos/eensymachines.in/luminapi
WORKDIR $HOME/repos/eensymachines.in/luminapi
COPY . .
RUN go mod download 
RUN go build -o luminapi .
