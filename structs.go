package denoue

import "fmt"

type JObject interface {
	GetKey() string
	fmt.Stringer
}

// JDict objects are maps wrapped by curly braces.
type JDict struct {
	objects map[string]JObject
}

func NewJDict() JDict {
	return JDict{
		objects: make(map[string]JObject),
	}
}

func (d *JDict) Set(elem JObject) {
	d.objects[elem.GetKey()] = elem
}

func (d *JDict) SetPair(key, val string) {
	pair := JPair{Key: key, Val: val}
	d.objects[pair.GetKey()] = pair
}

type JGroup struct {
	Key  string
	Dict JDict
}

func NewJGroup(key string, dict JDict) JGroup {
	return JGroup{Key: key, Dict: dict}
}

func (g JGroup) GetKey() string {
	return g.Key
}

// JArray objects have a key, and a list of values wrapped by square braces.
// JArray values can only be strings.
// JArrays can only be appended to, not changed.
type JArray struct {
	Key      string
	Vals     []string
	ByteVals [][]byte
}

func NewJArray(key string) JArray {
	return JArray{Key: key}
}

func (a JArray) GetKey() string {
	return a.Key
}

type escBuf []byte

func (b *escBuf) WriteEscaped(s string) {
	for i := 0; i < len(s); i++ {
		if s[i] == '"' {
			*b = append(*b, '\\')
		}
		*b = append(*b, s[i])
	}
}

func MakeSafe(s string) string {
	buf := make(escBuf, 0, 2*len(s))
	buf.WriteEscaped(s)
	return string(buf)
}

// AddSafe adds a formatted string and arguments to the JArray, escaping quotes.
func (a *JArray) AddSafe(format string, args ...string) {
	buf := make(escBuf, 0, 2*(len(format)+len(args)))
	buf.WriteEscaped(format)
	for _, arg := range args {
		buf.WriteEscaped(arg)
	}
	a.ByteVals = append(a.ByteVals, buf)
}

// Add adds a raw string to the JArray (no escaping).
func (a *JArray) Add(val string) {
	a.Vals = append(a.Vals, val)
}

// JPair is a key/value pair.
type JPair struct {
	Key, Val string
}

func NewJPair(key, val string) JPair {
	return JPair{Key: key, Val: val}
}

func (j JPair) GetKey() string {
	return j.Key
}
