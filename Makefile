setup-db:
	docker run --name chat-room -e POSTGRES_PASSWORD=123  -d  -e POSTGRES_DB=chat-room --network=chat-room  --expose=5432 postgres

setup-rabbitMQ:
	docker run --name rabbitMQ -d --network=chat-room -p 5672:5672 -p 15672:15672 rabbitmq:3-management