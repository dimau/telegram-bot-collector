# Telegram Bot Collector

Service is intended to collect all messages from all Telegram channels to which the user is subscribed and publish them all to the RabbitMQ queue

## How to run service

1. Build docker image
```
docker build --network host --build-arg TD_TAG=v1.8.0 --tag telegram-bot-collector .
```
2. Run docker container
```
docker run --rm -it -e "API_ID=..." -e "API_HASH=..." telegram-bot-collector
```