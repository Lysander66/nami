package main

import (
	"log"

	"github.com/opreader/nami"
	"github.com/opreader/nami/_examples/city/parser"
	"github.com/opreader/nami/_examples/city/pipeline"
)

func main() {
	itemPipeline, err := pipeline.ItemPipeline()
	if err != nil {
		log.Fatal(err)
	}
	e := nami.NewEngine(
		nami.WithProxy(proxies...),
		nami.WithQPS(2),
		nami.WithWorkerNum(3),
		nami.WithMaxDepth(3),
		nami.WithPipeline(itemPipeline),
	)
	url := `http://www.stats.gov.cn/tjsj/tjbz/tjyqhdmhcxhfdm/2020`
	task := nami.NewTask(url, parser.ParseProvince)
	e.Run(task)
}

var (
	proxies = []string{
		"http://136.243.254.196",
		"http://185.160.227.134",
		"http://58.220.95.90:9401",
	}
)
