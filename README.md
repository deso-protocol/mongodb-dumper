# mongodb-dumper

`mongodb-dumper` runs a full DeSo node and dumps the chain data into a MongoDB database

## Build

Running the following commands will create a Docker image called `mongodb-dumper:latest`.

1. Checkout `mongodb-dumper` and `core` in the same directory

2. In the `mongodb-dumper` repo, run the following (you may need sudo):

```
docker build -t mongodb-dumper -f Dockerfile ..
```

### Run

You may need sudo:

```
docker run -it mongodb-dumper /deso/bin/mongodb-dumper run
```

Configure the connection to mongodb:

```
   --mongodb-collection   string    MongoDB collection name  (default "data")
   --mongodb-database     string    MongoDB database name    (default "deso")
   --mongodb-uri          string    MongoDB connection URI   (default "mongodb://localhost:27017")
```

You may need to connect to the localhost network or supply DB authentication:

```
docker run --network="host" -it mongodb-dumper /deso/bin/mongodb-dumper run --mongo-uri "mongodb://userx:passwd@localhost:27017"
```

