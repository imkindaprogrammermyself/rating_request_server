## Personal Rating Server

A server written in `Go` as an Man-In-The-Middle for requesting player statistics.


### Why not just request the statistics directly? 

This server is intended to be used by multiple clients under the same rate limit. 

If multiple clients requests under the same rate limit, the limit will be exceeded and the API will return an error (`REQUEST_LIMIT_EXCEEDED`). 

This server will circumvent that since all connected clients will be on the same rate limit.

### Preparing the `.env` file.

The `.env` file should contain the following:

```
APPLICATION_ID=YOUR_APPLICATION_ID
SERVER_HOST=0.0.0.0
SERVER_PORT=35777
SERVER_TYPE=tcp
RATE_LIMIT=YOUR_APPLICATION_ID_RATE_LIMIT
```

 - You can get your own `APPLICATION_ID` at [https://developers.wargaming.net/applications/](https://developers.wargaming.net/applications/)
 - `RATE_LIMIT` is tied to your `APPLICATION_ID` see [docs](https://developers.wargaming.net/documentation/guide/principles/#application_types:~:text=to%20third%20parties.-,LIMITATIONS,-To%20provide%20the).

### Acquiring ship `expected values`

You can acquire the ships' expected values at [wows-numbers](https://na.wows-numbers.com/personal/rating)
Expected values and personal rating formula is acquired from [wows-numbers](https://na.wows-numbers.com/)


### Running the server

Install `Go` first.

```
go run main.go
```
You can specify the `.env` file or `expected.json` file path by supplying the flags `-env` (`.env` file path) or `-ev` (`expected.json` file path).

If those flags aren't supplied, the default path will be used i.e. the same location as the `main.go`

### Request format

The request should contain the size of the request + the request itself (in JSON format)

```
[request JSON size (4 bytes)][JSON data]
```
The JSON data format should be:

```
[{
	"id": 0,
	"realm": "ASIA",
	"account_id": 2000000001
}, {
	"id": 1,
	"realm": "ASIA",
	"account_id": 2000000002
}]
```

 - `id` can be anything unique and it must be a number.
 - `realm` should be `ASIA`, `NA`, `EU`, `RU` and it's case-sensitive.
 - `account_id` should be in the `realm`.

### Response Format

The response format is response size + response JSON

```
[response JSON size (4 bytes)][JSON data]
```
The response JSON format will be:
```
[{
	"id": 0,
	"rating": 9999.0
}, {
	"id": 1,
	"rating": 9999.0
}]
```
