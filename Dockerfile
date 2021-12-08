FROM alpine:edge AS mongodb-dumper

RUN apk update
RUN apk upgrade
RUN apk add --update go gcc g++ vips-dev

WORKDIR /deso/src

COPY mongodb-dumper/go.mod mongodb-dumper/
COPY mongodb-dumper/go.sum mongodb-dumper/
COPY core/go.mod core/
COPY core/go.sum core/

WORKDIR /deso/src/mongodb-dumper

RUN go mod download

# include mongodb-dumper src
COPY mongodb-dumper/cmd     cmd
COPY mongodb-dumper/mongodb mongodb
COPY mongodb-dumper/main.go .

# include core src
COPY core/desohash ../core/desohash
COPY core/cmd       ../core/cmd
COPY core/lib       ../core/lib
COPY core/migrate   ../core/migrate

# build mongodb-dumper
RUN GOOS=linux go build -mod=mod -a -installsuffix cgo -o bin/mongodb-dumper main.go

# create tiny image
FROM alpine:edge

COPY --from=mongodb-dumper /deso/src/mongodb-dumper/bin/mongodb-dumper /deso/bin/mongodb-dumper
