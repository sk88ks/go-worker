go-workers
====

Go-workers is a helper allow you to handle consistency process with goroutines

Installation
----

```
go get github.com/sk88ks/go-parse
```

Quick start
----

To create a session from default client,

```go
import(
  "github.com/sk88ks/go-wokers"
)

func main() {

  m := workers

}
```

Custom Client
----

To create a configured` client,

```go
parseClient := goparse.NewClient(goparse.Config({
  ApplicationId: "PARSE_APPLICATION_ID",
  RestAPIKey: "PARSE_REST_API_KEY",
  MasterKey: "PARSE_MASTER_KEY",
  EndPointURL: "PARSE_ENDPOINT_URL"
})

parseSession := parseClient.NewSession()
me, err := parseSession.GetMe()
..
```

Environment variables
----

The default client uses environment variables to access Parse REST API.

- `PARSE_APPLICATION_ID`
- `PARSE_REST_API_KEY`
- `PARSE_MASTER_KEY`
- `PARSE_ENDPOINT_URL`

License
----
Goparse is licensed under the MIT.
