# API server

REST API entry-point to the Emeris back-end.  
At its simplest the api-server is a translation layer between JSON REST and chain-specific RPC.  

## Dependencies & Licenses

The list of non-{Cosmos, AiB, Tendermint} dependencies and their licenses are:

|Module   	                  |License          |
|---	                      |---  	        |
|gin-gonic/gin   	          |MIT   	        |
|go-playground/validator   	  |MIT   	        |
|jmoiron/sqlx   	          |MIT   	        |
|go.uber.org/zap   	          |MIT           	|
|sigs.k8s.io/controller-runtime |MIT            |
|sony/sonyflake               |MIT              |
|lib/pq                       |Open use         |
