# Day-Trader

## Installation

Clone the repository
```bash
$ https://github.com/bhavanvir/day-trader.git
```
### Local setup using Docker
- First, install [docker](https://docs.docker.com/get-docker/) and [docker-compose](https://docs.docker.com/compose/install/).
```bash
$ cd Microservice1
```
```bash
$ docker compose up
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