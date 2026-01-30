# scalabe-coupon-excercise

Personal excercise for trying loadtesting, concurrency, lockings, loadtests, and maybe other stuff

Yes, I typo'd scalable

Currently, this implementation doesn't enforce strict FIFO.

## Tech Stack

- Golang
- PostgreSQL
- Redis
- Docker
- Locust

It's using gin, I'm aware is a bloated library, but it still works for now.

I'm aware the codebase is not clean.

## Prerequisites

- Docker
- Docker Compose

Should be enough to seamlessly run, assuming you don't use an ancient OS that doesn't support docker compose v2

## How to run

- docker-compose up --build

Wait for all the services to run correctly, should take about a minute.

WARNING: each time the go service starts, it will nuke and rebuild the tables, this is for ease of testing. Modify db.go if you don't want that behavior

Server runs on localhost:8080, db runs on localhost:5432, redis runs on localhost:6379

Then, there are three locust instances:

- localhost:8089 -> flash sale scenario
- localhost:8090 -> double dip scenario
- localhost:8091 -> personal sandbox

## How to test

### Flash Sale Scenario

After running docker compose, open localhost:8089. The tests are already setup, you'd only need to press "START".

You can see the test running on both the web ui and docker compose CLI.

Then you can see the final coupon information on the logs. Go to the logs tab and refresh the website. You should only see 5 ids at `claimed_by`

### Flash Sale Scenario

After running docker compose, open localhost:8090. The tests are already setup, you'd only need to press "START".

You can see the test running on both the web ui and docker compose CLI.

Then you can see the final coupon information on the logs. Go to the logs tab and refresh the website. You should only see 1 ids at `claimed_by`
