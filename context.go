package nami

import (
	"sync"
)

// Context provides a tiny layer for passing data between callbacks
type Context struct {
	MaxDepth   int
	contextMap map[string]interface{}
	lock       *sync.RWMutex
}

// NewContext initializes a new Context instance
func NewContext() *Context {
	return &Context{
		contextMap: make(map[string]interface{}),
		lock:       &sync.RWMutex{},
	}
}

// Put stores a value of any type in Context
func (c *Context) Put(key string, value interface{}) {
	c.lock.Lock()
	c.contextMap[key] = value
	c.lock.Unlock()
}

// Get retrieves a string value from Context.
// Get returns an empty string if key not found
func (c *Context) Get(key string) string {
	c.lock.RLock()
	defer c.lock.RUnlock()
	if v, ok := c.contextMap[key]; ok {
		return v.(string)
	}
	return ""
}

// GetAny retrieves a value from Context.
// GetAny returns nil if key not found
func (c *Context) GetAny(key string) interface{} {
	c.lock.RLock()
	defer c.lock.RUnlock()
	if v, ok := c.contextMap[key]; ok {
		return v
	}
	return nil
}

// ForEach iterate context
func (c *Context) ForEach(fn func(k string, v interface{}) interface{}) []interface{} {
	c.lock.RLock()
	defer c.lock.RUnlock()
	ret := make([]interface{}, 0, len(c.contextMap))
	for k, v := range c.contextMap {
		ret = append(ret, fn(k, v))
	}
	return ret
}

type Request struct {
	Ctx   *Context
	Url   string
	Depth int
}

type Response struct {
	Ctx     *Context
	Body    []byte
	Request *Request
}

type Task struct {
	Request *Request
	ParseFunc
}

func NewTask(url string, parser ParseFunc) Task {
	return Task{
		Request: &Request{
			Ctx:   NewContext(),
			Url:   url,
			Depth: 1,
		},
		ParseFunc: parser,
	}
}

type ParseFunc func(Response) Result

func NilParser(_ Response) Result {
	return Result{}
}

type Result struct {
	Tasks []Task
	Items []Item
}

func (r *Result) AddTask(url string, parser ParseFunc, ctxMap ...map[string]interface{}) {
	task := Task{
		Request:   &Request{Url: url},
		ParseFunc: parser,
	}
	if len(ctxMap) > 0 {
		ctx := NewContext()
		for k, v := range ctxMap[0] {
			ctx.Put(k, v)
		}
		task.Request.Ctx = ctx
	}
	r.Tasks = append(r.Tasks, task)
}

type Item struct {
	Id   string
	Url  string
	Data interface{}
}
