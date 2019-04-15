Ada Backend Challenge
==========

> How to easily get this program up and running on your computer using `go`, `docker` and `docker-compose`

## Prerequisites:

- Clone our API repo: `git clone -o origin git@github.com:kzap/ada-backend-challenge.git`

- Install [Go](https://golang.org/doc/install) for your Operating System.

- Install [Docker and Docker Compose](https://www.docker.com/community-edition) for your Operating System.

## How to run using Go:

1. Go to the `ada-backend-challenge` directory and run the main executable using `go run`.

  ```sh
  $ go run ./cmd/backend-challenge/
  ```

2. Open `http://0.0.0.0:8000` in your web browser to see the Index page.

## How to run using Docker:

1. In the `ada-backend-challenge` repo directory, copy `./deployments/docker-compose/.env.sample` to `./.env` and customize as needed.

2. Build and run the stack using `docker-compose`:

  ```sh
  docker-compose up --build -d
  ```

> Congratulations, you may now access API at `http://0.0.0.0:18080/` (or whatever your Docker Machine IP is and the port you specified in `ADA_WEB_HTTP_PORT`).

## Additional Tools:

We have additional `docker-compose` files you may use to enhance your development:

### MySQL Debugging

> Expose the internal database to your computer or access it via phpMyAdmin so you can view the database.

- Configure the following variables in your `.env` file:
  - `ADA_DB_MYSQL_PORT` - the port on your computer where you want the MySQL service inside the container to be available on.
  - `ADA_DBADMIN_PORT` - the port on your computer where you want phpMyAdmin to be accessible on.

- Run docker-compose with the following command
```sh
docker-compose -f docker-compose.yml -f deployments/docker-compose/docker-compose.dbdev.yml up
```

- Access phpMyAdmin at `http://0.0.0.0:18081` (or whatever port you put in `ADA_DBADMIN_PORT`).
