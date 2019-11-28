FROM golang:latest
WORKDIR /app
COPY . .
RUN go build -o main .
EXPOSE 4500
CMD ["./main"]

#sudo docker build -t  goserver:0.0-1 .