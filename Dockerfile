FROM kneerunjun/gogingonic:latest
# from the vanilla image of go gin with mgo driver

# mapping for log files

COPY . .
RUN go mod download 
RUN go build -o luminapi .