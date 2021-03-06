go-workers
====

[![Build Status](https://travis-ci.org/sk88ks/go-worker.svg?branch=master)](https://travis-ci.org/sk88ks/go-worker)
[![Coverage Status](https://coveralls.io/repos/sk88ks/go-worker/badge.svg?branch=master)](https://coveralls.io/r/sk88ks/go-worker?branch=master)

Go-workers is a helper allow you to handle consistency process with goroutines

Current API Documents:

* go-worker: [![GoDoc](https://godoc.org/github.com/sk88ks/go-worker?status.svg)](https://godoc.org/github.com/sk88ks/go-worker)

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

type Result struct {
  Sample1 String
  Sample2 int
}

func main() {

  var str string
  var i int
  workerNum := runtime.NumCPU()
  m := worker.New(workerNum)
  m.Add("Sample1", function1, "this", "is", "test")
  m.Add("Sample2", function2, 1, 2)
  m.Success(func(p *worker.Process) {
    // Process for success
    // Can add new worker process like
    // m.Add("success", funcWithSuccess, p.Result)
    if p.ID == "no1" {
      str = p.Result[0].(string)
    }
    
  })
  m.Fail(func(p *worker.Process) {
    // Precess for fail
    // Can stop all worker and process with Stop()
    // m.Stop()
  })
  
  // Also be able to retrieve results by using pointer struct
  result := Result{
  m.Run(&result)

}
```
