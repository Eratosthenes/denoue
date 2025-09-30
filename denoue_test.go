package denoue

import (
	"errors"
	"fmt"
	"testing"
)

// Benchmark_Printf-8         20131             97279 ns/op             112 B/op          1 allocs/op
func Benchmark_Printf(b *testing.B) {
	str := `{"time": "2023-12-15 7:26:25.600am -05", "level": "INFO", "msgs": ["Benchmarking JLog: hello world"]}`
	str += "\n"
	for i := 0; i < b.N; i++ {
		fmt.Print(str)
	}
}

// Benchmark_Printf2-8        28598             58689 ns/op             160 B/op         10 allocs/op
func Benchmark_Printf2(b *testing.B) {
	str := `{"time": "2023-12-15 7:26:25.600am -05", "level": "INFO", "msgs": ["Benchmarking JLog: hello world"]}`
	str += "\n"
	n := len(str)
	for i := 0; i < b.N; i++ {
		for j := 0; j < n; j += 10 {
			if j+10 >= n {
				fmt.Printf(str[j : j+n%10])
			} else {
				fmt.Print(str[j : j+10])
			}
		}
	}
}

// BenchmarkJLog_NewInfoAndPrint-8            32074             49843 ns/op            1704 B/op         51 allocs/op
func BenchmarkJLog_NewInfoAndPrint(b *testing.B) {
	for i := 0; i < b.N; i++ {
		jlog := New()
		jlog.Info("Benchmarking JLog: hello world")
		jlog.Print()
	}
}

// 2.085 ns/op	       0 B/op	       0 allocs/op
func BenchmarkJLog_Print(b *testing.B) {
	jlog := New()
	jlog.Info("Benchmarking JLog: hello world")
	for i := 0; i < b.N; i++ {
		jlog.Print()
	}
}

// 77.75 ns/op	      84 B/op	       0 allocs/op
func BenchmarkJLog_InfoAndPrint(b *testing.B) {
	jlog := New()
	for i := 0; i < b.N; i++ {
		jlog.Info("Benchmarking JLog: hello world")
		jlog.Print()
	}
}

// 179.2 ns/op	     185 B/op	       1 allocs/op
func BenchmarkJLog_InfoWithArgsAndPrint(b *testing.B) {
	jlog := New()
	for i := 0; i < b.N; i++ {
		jlog.Info("Benchmarking JLog: %s world", "hello")
		jlog.Print()
	}
}

func Test_InfoAndPrint(t *testing.T) {
	jlog := New()
	jlog.Info("Benchmarking JLog: hello world")
	jlog.Print()
}

func Test_NoPrint(t *testing.T) {
	jlog := New()
	jlog.Print()
}

func ExampleJArray_String() {
	a := NewJArray("array")
	a.Add("hello")
	a.Add("world")
	fmt.Println(a)

	// Output:
	// "array": ["hello", "world"]
}

func ExampleJDict_String() {
	d := NewJDict()
	d.SetPair("cat", "meow")
	d.SetPair("dog", "woof")
	fmt.Println(d)

	// Output:
	// {"cat": "meow", "dog": "woof"}
}

func TestJArray_Get(t *testing.T) {
	a := NewJArray("array")
	a.Add("hello")
	a.Add("world")

	jlog := New()
	jlog.Set(a)

	got_a, _ := Get[JArray](jlog, "array")
	got_a.Add("another greeting")
	jlog.Set(got_a)
	jlog.PrettyPrint()
}

func ExampleJGroup_String() {
	jlog := New()
	jlog.SetPair(TIME_KEY, "2023-08-11 10:22:59.654am -04")

	dict := NewJDict()
	dict.SetPair("remote_ip", "192.0.2.1")
	dict.SetPair("method", "GET")
	dict.SetPair("url", "/ping")

	group := JGroup{
		Key:  "request",
		Dict: dict,
	}

	jlog.Set(group)
	jlog.PrettyPrint()

	// Output:
	// {
	//   "time": "2023-08-11 10:22:59.654am -04",
	//   "level": "INFO",
	//   "request": {
	//     "method": "GET",
	//     "remote_ip": "192.0.2.1",
	//     "url": "/ping"
	//   }
	// }
}

func ExampleJLog_PrettyPrint() {
	jlog := New()
	// NOTE: generally the time doesn't have to be explicitly set,
	// but we set it explicitly here so that the test will pass
	jlog.SetPair(TIME_KEY, "2023-08-10 9:00:41.553pm -04")

	dict := NewJDict()
	dict.SetPair("url", "/ping")
	dict.SetPair("remote_ip", "192.0.2.1")
	dict.SetPair("method", "GET")

	group := JGroup{
		Key:  "request",
		Dict: dict,
	}
	jlog.Set(group)
	jlog.SetPair("response", "200 OK")

	jlog.Info("some info")
	jlog.Info("some more info")
	jlog.PrettyPrint()

	// Output:
	// {
	//   "time": "2023-08-10 9:00:41.553pm -04",
	//   "level": "INFO",
	//   "msgs": [
	//     "some info",
	//     "some more info"
	//   ],
	//   "request": {
	//     "method": "GET",
	//     "remote_ip": "192.0.2.1",
	//     "url": "/ping"
	//   },
	//   "response": "200 OK"
	// }
}

func ExampleJLog_Warn() {
	jlog := New()
	jlog.SetPair(TIME_KEY, "2023-08-10 9:16:35.569pm -04")

	dict := NewJDict()
	dict.SetPair("url", "/ping")
	dict.SetPair("remote_ip", "192.0.2.1")
	dict.SetPair("method", "GET")

	group := JGroup{
		Key:  "request",
		Dict: dict,
	}
	jlog.Set(group)

	jlog.SetPair("response", "200 OK")

	jlog.Info("some info")
	jlog.Warn("this is a warning")
	jlog.Info("some more info")
	jlog.PrettyPrint()

	// Output:
	// {
	//   "time": "2023-08-10 9:16:35.569pm -04",
	//   "level": "WARN",
	//   "msgs": [
	//     "some info",
	//     "this is a warning",
	//     "some more info"
	//   ],
	//   "request": {
	//     "method": "GET",
	//     "remote_ip": "192.0.2.1",
	//     "url": "/ping"
	//   },
	//   "response": "200 OK"
	// }
}

func ExampleJLog_Error() {
	jlog := New()
	jlog.SetPair(TIME_KEY, "2023-08-10 9:29:05.722pm -04")

	jlog.Info("some info")
	jlog.Warn("this is a warning")
	jlog.Error(errors.New("some error"))
	jlog.PrettyPrint()

	// Output:
	// {
	//   "time": "2023-08-10 9:29:05.722pm -04",
	//   "level": "ERROR",
	//   "error": "some error",
	//   "msgs": [
	//     "some info",
	//     "this is a warning"
	//   ]
	// }
}

func ExampleJLog_PrettyPrint_subgroup() {
	jlog := New()
	jlog.SetPair(TIME_KEY, "2023-08-10 9:17:43.348pm -04")

	req := NewJDict()
	req.SetPair("url", "/ping")
	req.SetPair("remote_ip", "192.0.2.1")
	req.SetPair("method", "GET")

	src := NewJDict()
	src.SetPair("source", "program/file:32")
	req.Set(NewJGroup("caller", src))

	jlog.Set(NewJGroup("request", req))
	jlog.Info("some info")
	jlog.Info("some more info")
	jlog.Error(errors.New("this is an error"))
	jlog.PrettyPrint()

	// Output:
	// {
	//   "time": "2023-08-10 9:17:43.348pm -04",
	//   "level": "ERROR",
	//   "error": "this is an error",
	//   "msgs": [
	//     "some info",
	//     "some more info"
	//   ],
	//   "request": {
	//     "caller": {
	//       "source": "program/file:32"
	//     },
	//     "method": "GET",
	//     "remote_ip": "192.0.2.1",
	//     "url": "/ping"
	//   }
	// }

}

func ExampleJLog_PrettyPrint_change_pairs() {
	jlog := New()
	jlog.SetPair(TIME_KEY, "2023-08-10 9:20:27.067pm -04")

	dict := NewJDict()
	dict.SetPair("url", "/ping")
	dict.SetPair("remote_ip", "192.0.2.1")
	dict.SetPair("method", "GET")

	group := NewJGroup("request", dict)
	jlog.Set(group)
	jlog.SetPair("response", "200 OK")
	jlog.SetPair("caller", "program/file:32")

	jlog.Info("generic info")
	jlog.Warn("this is a warning")

	// let's change some pairs
	newResponse := NewJPair("response", "404 Not Found")
	jlog.Set(newResponse)
	jlog.SetPair("caller", "program/file:64")

	jlog.Error(errors.New("this is an error"))
	jlog.Info("more generic info")
	jlog.PrettyPrint()

	// Output:
	// {
	//   "time": "2023-08-10 9:20:27.067pm -04",
	//   "level": "ERROR",
	//   "error": "this is an error",
	//   "caller": "program/file:64",
	//   "msgs": [
	//     "generic info",
	//     "this is a warning",
	//     "more generic info"
	//   ],
	//   "request": {
	//     "method": "GET",
	//     "remote_ip": "192.0.2.1",
	//     "url": "/ping"
	//   },
	//   "response": "404 Not Found"
	// }
}

func ExampleJLog_Log() {
	var errTest = errors.New("test error")
	jlog := New()
	jlog.SetPair(TIME_KEY, "2023-08-10 9:20:27.067pm -04")
	jlog.SetPair("caller", "program/file:32")

	jlog.Info("some info")
	jlog.Error(errors.New("this is an error"))

	f := func(err error, args ...string) (string, []string, []JObject) {
		var level string
		if errors.Is(err, errTest) {
			level = WARN
		}

		var objs []JObject
		dict := NewJDict()
		dict.SetPair("url", "/ping")
		dict.SetPair("remote_ip", "192.0.2.1")
		dict.SetPair("method", "GET")

		group := NewJGroup("request", dict)
		objs = append(objs, group)

		pair := JPair{Key: "error", Val: err.Error()}
		objs = append(objs, pair)

		return level, []string{}, objs
	}

	jlog.Log(f, errTest)
	jlog.PrettyPrint()

	// Output:
	// {
	//   "time": "2023-08-10 9:20:27.067pm -04",
	//   "level": "WARN",
	//   "error": "test error",
	//   "caller": "program/file:32",
	//   "msgs": [
	//     "some info"
	//   ],
	//   "request": {
	//     "method": "GET",
	//     "remote_ip": "192.0.2.1",
	//     "url": "/ping"
	//   }
	// }
}
