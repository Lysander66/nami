package nami

import (
	"bufio"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/opreader/nami/common"
	"github.com/opreader/nami/proxy"
	"golang.org/x/net/html/charset"
	"golang.org/x/text/transform"
)

type Engine struct {
	Pipeline           chan Item
	Ctx                *Context
	scheduler          Scheduler
	workerNum          int
	QPS                int
	rateLimiter        <-chan time.Time
	randomUserAgent    bool
	useProxy           bool
	roundRobinSwitcher *proxy.RoundRobinSwitcher
	wg                 *sync.WaitGroup
	lock               *sync.RWMutex
}

func NewEngine(options ...Option) *Engine {
	e := &Engine{
		Ctx:             NewContext(),
		scheduler:       NewTaskScheduler(),
		randomUserAgent: true,
	}
	for _, opt := range options {
		opt(e)
	}
	return e
}

type Option func(*Engine)

func WithScheduler(scheduler Scheduler) Option {
	return func(e *Engine) {
		e.scheduler = scheduler
	}
}

func WithWorkerNum(num int) Option {
	return func(e *Engine) {
		e.workerNum = num
	}
}

func WithMaxDepth(maxDepth int) Option {
	return func(e *Engine) {
		e.Ctx.MaxDepth = maxDepth
	}
}

func WithQPS(QPS int) Option {
	return func(e *Engine) {
		e.QPS = QPS
		e.rateLimiter = time.Tick(time.Second / time.Duration(QPS))
	}
}

func WithProxy(ProxyURLs ...string) Option {
	return func(e *Engine) {
		switcher, err := proxy.RoundRobinProxySwitcher(ProxyURLs...)
		if err != nil {
			log.Fatal(err)
		}
		e.roundRobinSwitcher = switcher
		e.useProxy = true
	}
}

func WithRandomUserAgent(randomUserAgent bool) Option {
	return func(e *Engine) {
		e.randomUserAgent = randomUserAgent
	}
}

func WithPipeline(pipeline chan Item) Option {
	return func(e *Engine) {
		e.Pipeline = pipeline
	}
}

func (e *Engine) Run(tasks ...Task) {
	out := make(chan Result)
	e.scheduler.Run()
	for i := 0; i < e.workerNum; i++ {
		e.createWorker(e.scheduler.Worker(), out)
	}
	for _, task := range tasks {
		e.scheduler.Submit(task)
	}
	for {
		result := <-out
		for _, item := range result.Items {
			go func(item Item) {
				e.Pipeline <- item
			}(item)
		}
		for _, task := range result.Tasks {
			if !isDuplicate(task.Request.Url) {
				e.scheduler.Submit(task)
			}
		}
	}
}

func (e *Engine) createWorker(in chan Task, out chan<- Result) {
	go func() {
		for {
			e.scheduler.WorkerReady(in)
			task := <-in
			result, err := e.Process(task)
			if err != nil {
				//e.scheduler.Submit(task) //retry
				continue
			}
			out <- result
		}
	}()
}

func (e *Engine) Process(task Task) (Result, error) {
	result := Result{}
	depth := task.Request.Depth
	if e.Ctx.MaxDepth > 0 && depth > e.Ctx.MaxDepth {
		return result, common.ErrMaxDepth
	}
	bytes, err := e.fetch(task.Request.Url)
	if err != nil {
		return result, err
	}
	// Ctx is a context between a Request and a Response
	resp := Response{
		Ctx:     e.Ctx,
		Body:    bytes,
		Request: task.Request,
	}
	result = task.ParseFunc(resp)
	next := depth + 1
	for _, task := range result.Tasks {
		task.Request.Depth = next
	}
	return result, nil
}

// Todo: no need to create a new client object every time
func (e *Engine) fetch(Url string) ([]byte, error) {
	if e.QPS > 0 {
		<-e.rateLimiter
	}
	c := &http.Client{}
	if e.useProxy {
		proxyURL := e.roundRobinSwitcher.GetProxy()
		c.Transport = &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}
	}
	req, err := http.NewRequest(http.MethodGet, Url, nil)
	if err != nil {
		return nil, err
	}
	if e.randomUserAgent {
		req.Header.Add("User-Agent", proxy.RandomUserAgent())
	}
	resp, err := c.Do(req)
	if err != nil {
		log.Printf("fetch err: %v\n", err)
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(http.StatusText(resp.StatusCode))
	}
	// DetermineEncoding
	r := bufio.NewReader(resp.Body)
	peek, err := r.Peek(1024)
	if err != nil {
		return nil, err
	}
	encoding, _, _ := charset.DetermineEncoding(peek, "")
	reader := transform.NewReader(r, encoding.NewDecoder())
	return ioutil.ReadAll(reader)
}

var visitedUrls = make(map[string]bool)

func isDuplicate(url string) bool {
	if visitedUrls[url] {
		return true
	}
	visitedUrls[url] = true
	return false
}
