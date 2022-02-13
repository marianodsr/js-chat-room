setup-db:
	docker run --name chat-room -e POSTGRES_PASSWORD=123 -p 5432:5432 -d  -e POSTGRES_DB=chat-room postgres