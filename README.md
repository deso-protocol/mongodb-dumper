# mongodb-dumper

`mongodb-dumper` runs a full BitClout node and dumps the chain data into a MongoDB database

## Build

Running the following commands will create a Docker image called `mongodb-dumper:latest`.

1. Checkout `mongodb-dumper` and `core` in the same directory

2. In the `mongodb-dumper` repo, run the following (you may need sudo):

```
docker build -t mongodb-dumper -f Dockerfile ..
```


### Setup MongoDB Server

There are different ways how you could setup a mongoDB server. You could either buy a managed mongoDB server (e.g. at [render.com](https://render.com/docs/deploy-mongodb) ), run a mongoDB server directly on your host system or run a mongoDB server inside a separate docker container. The last option is recommended.

Here is a quick guide on how to run a mongoDB server in a separate docker container.

1. First create a new directory:
```
mkdir mongodb
```

Your directory structure should now look as following:
```
- mongodb
- mongodb-dumper
- core
```

2. Switch to the newly created `mongodb` directory:
```
cd mongodb
```

3. Create a new docker compose file to specifcy attributes of your mongodb docker container

docker-compose.yml
```
version: '3.7'
services:
  mongodb_container:
    image: mongo:latest
    environment:
      MONGO_INITDB_ROOT_USERNAME: root
      MONGO_INITDB_ROOT_PASSWORD: rootpassword
    ports:
      - 27017:27017
    volumes:
      - mongodb_data_container:/data/db

volumes:
  mongodb_data_container:
```

Your mongodb username will be `root` and your password will be `rootpassword`. At this point you should watch out if you expose the port `27017` of your host system to the internet. In case you do expose this port, you should change the username and the password to something else otherwise attackers from outside might delete your data or worse manipulate it without you knowing about it.

4. Run mongodb container
```
docker-compose up
```

### Run

You may need sudo:

```
docker run -it mongodb-dumper /bitclout/bin/mongodb-dumper run
```

Configure the connection to mongodb:

```
   --mongodb-collection   string    MongoDB collection name  (default "data")
   --mongodb-database     string    MongoDB database name    (default "bitclout")
   --mongodb-uri          string    MongoDB connection URI   (default "mongodb://localhost:27017")
```

You may need to connect to the localhost network or supply DB authentication:

```
docker run --network="host" -it mongodb-dumper /bitclout/bin/mongodb-dumper run --mongo-uri "mongodb://root:rootpassword@localhost:27017"
```

