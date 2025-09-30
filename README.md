# Denoue: A small package for JSON-structured logging

## About:
Denoue is a structured logging library for Go that defers logging until the end of request handling.

This package defines JObjects, which are objects that can be printed as JSON. See the examples section for examples of each printable type (JGroup, JDict, JArray, JPair).

---
## Documentation:
For documentation, run:
```
go install golang.org/x/pkgsite/cmd/pkgsite@latest
```

then run:
```
pkgsite >/dev/null 2>&1 &
```
and navigate to http://localhost:8080 in your browser.

## Features:
- Allows the printing of log statements to be deferred until after a request has finished processing (see usage). This allows us to log more information throughout the lifetime of a request without incurring a significant performance penalty.
- Allows log statements to be modified, removed, or appended to before they are printed.
- Logs are printed in valid JSON using a subset of the JSON grammar specification (only allowing strings as values).
- Ensures no duplicate keys are created.
- Safe for concurrent usage; every operation is wrapped in mutexes. Data race safety was tested by `go tool -race`.
- Includes an interface for generating mocks to allow logging functionality to be easily tested.
- Includes functionality for pretty-printing and printing multiple times per logger instance for debugging purposes.
- Code is reasonably optimized and performs well on benchmarks.
- Code base is small, manageable, and well-documented.
- Code test coverage is high.

## Design:

There are five data types defined in this library: JLog, JGroup, JDict, JArray, JPair. 

### JLog
JLog is a `denoue` instance. It is instantiated with the `New()` method:
```
jlog := New()
```

Alternatively, if `denoue` is being imported by another package, we can instantiate it this way: 

```
jlog := denoue.New()
```

Once the `denoue` is created, it can log to `os.Stdout` right away by default. To log to another output stream, the `SetOutput` method can optionally be called. For example:
```
jlog.SetOutput(os.Stderr)
```

The `Info` function prints the log to an internal buffer. In order to display the log on `os.Stdout`, the `Print` statement must be called. For example, the statement below on its own does not display anything:
```
jlog.Info("hello world")
```
However, if it is followed by:
```
jlog.Print()
```
This produces the following log:
```
{
  "time": "2023-12-21 2:58:44.083pm -05",
  "level": "INFO",
  "msgs": [
    "hello world"
  ]
}
```
Note that three default fields are created: `time`, `level`, and `msgs`. These three fields are always included in log statements.

Log levels follow a hierarchy with three levels: INFO, WARN, and ERROR, with error as the highest level. The default level is INFO. If a `Warn()` function is called at any point in the jlog's lifetime, then the level is promoted to WARN. For example:
```
jlog := New()
jlog.Info("hello world")
jlog.Warn("flash flooding until 10pm")
jlog.Print()
```

This code produces the following log:
```
{
  "time": "2023-12-21 3:08:14.034pm -05",
  "level": "WARN",
  "msgs": [
    "hello world",
    "flash flooding until 10pm"
  ]
}
```

Notice that the log level was promoted to WARN, and the WARN message was simply appended to the `msgs` field. Levels can only increase, never decrease. For example, calling `Info()` again adds another message to `msgs` but does not lower the log level:
```
jlog := New()
jlog.Info("hello world")
jlog.Warn("flash flooding until 10pm")
jlog.Info("hello again, world")
jlog.Print()
```

This produces:
```
{
  "time": "2023-12-21 3:11:18.516pm -05",
  "level": "WARN",
  "msgs": [
    "hello world",
    "flash flooding until 10pm",
    "hello again, world"
  ]
}
```

The `Error()` method takes an `error` type rather than a string, promotes the log level to ERROR, and adds a new field called `error` to the log. For example:
```
jlog := New()
jlog.Info("hello world")
jlog.Warn("flash flooding until 10pm")
jlog.Error(errors.New("this vehicle is not amphibious"))
jlog.Info("hello again, world")
jlog.Print()
```

This produces:
```
{
  "time": "2023-12-21 3:20:00.651pm -05",
  "level": "ERROR",
  "error": "this vehicle is not amphibious",
  "msgs": [
    "hello world",
    "flash flooding until 10pm",
    "hello again, world"
  ]
}
```

Log statements only contain one `"error"` field because it is assumed that only one error is called (any further errors should either wrap or overwrite the original error).

### JObject
A JObject is an object that implements the JObject interface. JObjects can be looked up by key, and can be printed. The following types (to be discussed below) implement the JObject interface: JPair, JArray, JGroup.

A logging instance can accept any JObject through its `Set()` method.

### JPair
A JPair is simply a key-value pair. A JPair can be created through the `SetPair` method, which can be called on either a JLog type or a JDict type. For example:
```
jlog := New()
jlog.Info("hello world")
jlog.SetPair("method", "GET")
jlog.Print()
```
This adds a new field `"method"` with a value `"GET"` to the log statement:
```
{
  "time": "2023-12-21 5:49:29.821pm -05",
  "level": "INFO",
  "method": "GET",
  "msgs": [
    "hello world"
  ]
}
```

Alternatively, the JPair could have been set this way:
```
jlog := New()
jlog.Info("hello world")

p := NewJPair("method", "GET")
jlog.Set(p)
jlog.Print()
```

However, the `SetPair()` method is slightly more convenient.

### JArray
A JArray is a key-value pair, where the value is an array. We can create an array and add it to a `JLog` using the `Set()` function as follows:

```
a := NewJArray("array")
a.Add("hello")
a.Add("world")

jlog := New()
jlog.Set(a)
jlog.Print()
```

This produces the log statement:
```
{
  "time": "2023-12-21 6:03:42.795pm -05",
  "level": "INFO",
  "array": [
    "hello",
    "world"
  ]
}
```

We can also retrieve the array later on and modify it, for example:
```
a := NewJArray("array")
a.Add("hello")
a.Add("world")

jlog := New()
jlog.Set(a)

got_a, _ := Get[JArray](jlog, "array")
got_a.Add("another greeting")
jlog.Set(got_a)
jlog.Print()
```

This produces the log:
```
{
  "time": "2023-12-21 6:11:55.387pm -05",
  "level": "INFO",
  "array": [
    "hello",
    "world",
    "another greeting"
  ]
}
```

We can also simply remove any JObject from a `denoue` using the `Pop` method:
```
a := NewJArray("array")
a.Add("hello")
a.Add("world")

jlog := New()
jlog.Set(a)

jlog.Pop("array")
jlog.Print()
```

This prints the log statement:
```
{
  "time": "2023-12-21 8:29:10.321pm -05",
  "level": "INFO"
}
```

### JDict
A JDict, unlike other types, does not have a key and does not implement the JObject interface. It holds a map of JObjects. JDict is not added to a denoue directly, rather, it is created and then added as part of a JGroup. We will see this in the next section. 

Here is an example of a JDict:
```
d := NewJDict()
d.SetPair("cat", "meow")
d.SetPair("dog", "woof")
fmt.Println(d)
```

This prints as:
```
{"cat": "meow", "dog": "woof"}
```

We can also add any type of JObject to a JDict, for example, to add a JArray:
```
d := NewJDict()
d.SetPair("cat", "meow")
d.SetPair("dog", "woof")

a := NewJArray("array")
a.Add("hello")
a.Add("world")

d.Set(a)
fmt.Println(d)
```
This prints as:
```
{"array": ["hello", "world"], "cat": "meow", "dog": "woof"}
```

### JGroup
A JGroup is a JObject that has a key and takes a JDict as its value. A denoue can take any JObject, so it can accept a JGroup. For example:

```
jlog := New()

d := NewJDict()
d.SetPair("cat", "meow")
d.SetPair("dog", "woof")

jlog.Set(NewJGroup("animals", d))
jlog.Print()
```

This prints the log statement:
```
{
  "time": "2023-12-21 7:42:04.966pm -05",
  "level": "INFO",
  "animals": {
    "cat": "meow",
    "dog": "woof"
  }
}
```

As before, the JGroup can be easily retrieved and modified. Continuing from the previous example:
```
g, _ := Get[JGroup](jlog, "animals")
g.Dict.SetPair("dog", "bark")
jlog.Set(g)
```

This prints the log statement:
```
{
  "time": "2023-12-21 7:45:43.561pm -05",
  "level": "INFO",
  "animals": {
    "cat": "meow",
    "dog": "bark"
  }
}
```

## Usage:

Here is a more detailed example of a `denoue` being instantiated in a middleware function called LogMiddleware:
```
type ReqID struct{}
var ReqIDKey ReqID

type denoueKey struct{}
var LogKey denoueKey

func LogMiddleware(h http.HandlerFunc) (http.HandlerFunc) {
	fn := func(w http.ResponseWriter, r *http.Request) {
		// create logger
		jlog := denoue.New()
		jlog.SetOutput(os.Stderr)

		// create request group
		reqDict := denoue.NewJDict()
		reqDict.SetPair("url", r.URL.Path)
		reqDict.SetPair("method", r.Method)
		reqDict.SetPair("remote_ip", r.RemoteAddr)

		remoteReqID := r.URL.Query().Get(REQUEST_ID)
		if remoteReqID != "" {
			reqDict.SetPair(REQUEST_ID, remoteReqID)
		}

		req := denoue.NewJGroup("request", reqDict)
		reqID := keys.GenerateUUID()
		jlog.SetPair(REQUEST_ID, reqID)
		jlog.Set(req)

		// add logger + request_id to context
		ctx := r.Context()
		ctx = context.WithValue(ctx, LogKey, jlog)
		ctx = context.WithValue(ctx, ReqIDKey, reqID)
		r = r.WithContext(ctx)

		// log request
		if _, found := RouteMap[r.URL.Path[1:]]; !found {
			jlog.Warn("unrecognized request received")
		} else {
			jlog.Info("request received")
		}

		h.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
```

Note that in the example above, there is no `defer jlog.Print()` method being called. Instead, we can call it in another middleware function that does basic request authorization and validation, eg:
```
// AuthMiddleware does basic authorization and request validation.
func AuthMiddleware(h http.HandlerFunc) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		// retrieve logger from request context
		jlog := r.Context().Value(LogKey).(*denoue.JLog)
		defer jlog.Print()

		// validate http method
		validMethods := RouteMap[r.URL.Path[1:]]
		if !validMethod(w, r, validMethods...) {
			jlog.Warn("invalid method")
			return
		}

		h.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
```

This method of injecting `denoue` into middleware works as long as the `denoue` is instantiated inside the first middleware function that is executed and printed from the last function. Alternatively, one could simply instantiate `denoue` and call `defer denoue.Print()` in the same middleware function; that works as well, but with this slightly more complicated example we can see how `denoue` is extracted from the request context, and we can also see how it can start logging from within the first middleware interceptor. 