// Package for JSON logging.
//
// This package defines JObjects, which are objects that can be printed as JSON. See the examples section for examples of each printable type (JGroup, JDict, JArray, JPair).
package denoue

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"sync"
	"time"
)

type JLogger interface {
	Get(k string) (JObject, error)
	Pop(k string) (JObject, error)
	SetPair(key, val string)
	Set(elem JObject)
	SetOutput(out io.Writer)
	SetTime(timeLayout string)
	Print()
	Reset()
	PrettyPrint()
	Info(format string, args ...string)
	Warn(format string, args ...string)
	Error(err error)
	Log(f LogFunc, err error, args ...string)
}

type JLog struct {
	out        io.Writer
	msgs       JArray
	level      string
	timeLayout string
	objects    map[string]JObject
	once       *sync.Once
	mu         *sync.Mutex
}

// New creates a new json logger.
func New() *JLog {
	return &JLog{
		out:        os.Stdout,
		msgs:       JArray{Key: MSG_KEY},
		level:      INFO,
		objects:    make(map[string]JObject),
		once:       new(sync.Once),
		mu:         new(sync.Mutex),
		timeLayout: DEFAULT_TIME_LAYOUT,
	}
}

// Get retrieves a particular JSON object.
func (j *JLog) Get(k string) (JObject, error) {
	j.mu.Lock()
	defer j.mu.Unlock()

	obj, ok := j.objects[k]
	if !ok {
		return nil, errors.New("key not found")
	}
	return obj, nil
}

type KeyVal interface {
	JPair | JArray | JGroup
}

// Get retrieves some JSON object that can be looked up by key.
func Get[T KeyVal](j *JLog, k string) (*T, error) {
	obj, found := j.objects[k]
	if !found {
		return nil, errors.New("key not found")
	}

	got, ok := obj.(T)
	if !ok {
		return nil, fmt.Errorf("could not cast object from key '%v'", k)
	}
	return &got, nil
}

// Pop retrieves and removes a particular JSON object.
func (j *JLog) Pop(k string) (JObject, error) {
	j.mu.Lock()
	defer j.mu.Unlock()

	obj, ok := j.objects[k]
	if !ok {
		return nil, errors.New("key not found")
	}
	delete(j.objects, k)
	return obj, nil
}

// Set a particular JSON object to another value for a given JPair.
func (j *JLog) SetPair(key, val string) {
	j.mu.Lock()
	defer j.mu.Unlock()

	pair := JPair{Key: key, Val: val}
	j.objects[pair.GetKey()] = pair
}

// Set JGroup, JPair, JDict, or JArray.
func (j *JLog) Set(elem JObject) {
	j.mu.Lock()
	defer j.mu.Unlock()

	j.objects[elem.GetKey()] = elem
}

// SetOutput sets the output.
func (j *JLog) SetOutput(out io.Writer) {
	j.mu.Lock()
	defer j.mu.Unlock()

	j.out = out
}

// SetTime sets a time layout.
func (j *JLog) SetTime(timeLayout string) {
	j.mu.Lock()
	defer j.mu.Unlock()

	j.timeLayout = timeLayout
}

// Reset allows you to print more than once (for debugging).
func (j *JLog) Reset() {
	j.once = new(sync.Once)
}

// Print the log statement (only once per request).
func (j *JLog) Print() {
	j.once.Do(func() {
		ts := time.Time.Format(time.Now(), j.timeLayout)
		dict := NewJDict()
		dict.SetPair(TIME_KEY, ts)
		dict.SetPair(LEVEL_KEY, j.level)
		if len(j.msgs.Vals) > 0 {
			dict.Set(j.msgs)
		} else {
			return // don't print if there are no messages
		}
		for _, elem := range j.objects {
			dict.Set(elem)
		}

		// instead of this:
		// fmt.Fprintf(j.out, "%v\n", dict)
		// we can go faster by writing directly
		var buf bytes.Buffer
		buf.WriteString(dict.String() + "\n")
		j.out.Write(buf.Bytes())
	})
}

// Pretty-print the log statement (only once per request).
// NOTE: This function is for debugging only. For production, use Print() instead.
func (j *JLog) PrettyPrint() {
	j.once.Do(func() {
		ts := time.Time.Format(time.Now(), j.timeLayout)
		dict := NewJDict()
		dict.SetPair(TIME_KEY, ts)
		dict.SetPair(LEVEL_KEY, j.level)
		if len(j.msgs.Vals) > 0 {
			dict.Set(j.msgs)
		}
		for _, elem := range j.objects {
			dict.Set(elem)
		}

		dir, _ := os.MkdirTemp("", "test_*")
		_ = os.WriteFile(dir+"/test.json", []byte(dict.String()), 0660)

		cmdStr := "cd %v && cat test.json | jq"
		cmd := fmt.Sprintf(cmdStr, dir)
		e := exec.Command("/bin/bash", "-c", cmd)
		e.Stdout = os.Stdout

		if err := e.Run(); err != nil {
			log.Fatalf("error: %v", err)
		}
	})
}

// Info level logging (doesn't print).
func (j *JLog) Info(format string, args ...string) {
	j.mu.Lock()
	defer j.mu.Unlock()

	if len(args) > 0 {
		j.msgs.AddSafe(format, args...)
	} else {
		j.msgs.Add(format)
	}
}

// Warn level logging (doesn't print).
func (j *JLog) Warn(format string, args ...string) {
	j.mu.Lock()
	defer j.mu.Unlock()

	if j.level != ERROR {
		j.level = WARN
	}

	if len(args) > 0 {
		j.msgs.AddSafe(format, args...)
	} else {
		j.msgs.Add(format)
	}
}

// Error level logging (doesn't print).
func (j *JLog) Error(err error) {
	j.mu.Lock()
	defer j.mu.Unlock()

	j.level = ERROR
	pair := JObject(JPair{Key: ERR_KEY, Val: MakeSafe(err.Error())})
	j.objects[pair.GetKey()] = pair
}

// LogFunc returns level, messages, objects
type LogFunc func(err error, args ...string) (string, []string, []JObject)

// Log executes a custom logging function.
func (j *JLog) Log(f LogFunc, err error, args ...string) {
	j.mu.Lock()
	defer j.mu.Unlock()

	level, msgs, objects := f(err, args...)
	j.level = level
	j.msgs.Vals = append(j.msgs.Vals, msgs...)

	for _, obj := range objects {
		j.objects[obj.GetKey()] = obj
	}
}
