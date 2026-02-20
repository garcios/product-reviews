
kill-subgraphs:
	lsof -i :4001 -t | xargs kill -9
	lsof -i :4002 -t | xargs kill -9
	lsof -i :4003 -t | xargs kill -9

kill-gateway:
	lsof -i :4000 -t | xargs kill -9

kill-api-user:
	lsof -i :8080 -t | xargs kill -9

kill-api-reviews:
	lsof -i :8081 -t | xargs kill -9

kill-api-products:
	lsof -i :8082 -t | xargs kill -9

kill-all:
	kill-subgraphs
	kill-gateway
	kill-api-user
	kill-api-reviews
	kill-api-products