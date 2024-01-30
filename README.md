# Day-Trader

## Installation

Clone the repository
```bash
$ https://github.com/bhavanvir/day-trader.git
```
### Local setup using Docker
- First, install [docker](https://docs.docker.com/get-docker/) and [docker-compose](https://docs.docker.com/compose/install/).
```bash
$ docker compose up --build
```

- In a new terminal
```bash
$ cd Client/app
```
```bash
$ npm i
```
```bash
$ npm start
```

- Finally, when you're done running the project run the following command to shut down the containers.

```bash
$ docker compose down
```

## API
To use createStock post request
```
$ curl -X POST -H "Content-Type: application/json" -d '{"stock_name": "ExampleStock"}' http://localhost:8080/createStock
```