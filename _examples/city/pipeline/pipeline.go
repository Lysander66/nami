package pipeline

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/opreader/nami"
	"github.com/opreader/nami/_examples/city/model"
)

const (
	cityIndex = "city"
	esNode    = "http://120.76.129.16:9200"
)

func ItemPipeline() (chan nami.Item, error) {
	client, err := elasticsearch.NewClient(
		elasticsearch.Config{
			Addresses: []string{esNode},
		})
	if err != nil {
		return nil, err
	}
	out := make(chan nami.Item)
	go func() {
		itemCount := 0
		for {
			itemCount++
			item := <-out
			err := Save(client, item)
			if err != nil {
				log.Printf("item pipeline: error saving item %v: %v", item, err)
			}
		}
	}()
	return out, nil
}

func Save(client *elasticsearch.Client, item nami.Item) error {
	city, ok := item.Data.(model.City)
	if !ok {
		log.Println("save error")
	}
	log.Println(city.Id, city.Name)

	b, err := json.Marshal(city)
	if err != nil {
		return err
	}
	req := esapi.IndexRequest{
		Index:      cityIndex,
		DocumentID: strconv.Itoa(city.Id),
		Body:       bytes.NewReader(b),
		Refresh:    "true",
	}
	resp, err := req.Do(context.Background(), client)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.IsError() {
		log.Printf("[%s] Error indexing document", resp.Status())
		return errors.New(fmt.Sprintf("[%s] Error indexing document", resp.Status()))
	}
	return nil
}
