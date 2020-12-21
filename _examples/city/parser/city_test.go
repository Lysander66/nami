package parser

import (
	"fmt"
	"io/ioutil"
	"log"
	"testing"

	"github.com/opreader/nami"
	"github.com/opreader/nami/_examples/city/model"
)

func TestParseProvince(t *testing.T) {
	filename := fmt.Sprintf(localPath, "country.html")
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}
	resp := nami.Response{
		Body:    bytes,
		Request: &nami.Request{Url: indexUrl},
	}
	result := ParseProvince(resp)
	if got := len(result.Items); got != 31 {
		t.Errorf("ParseCity() = %v, want %v", got, 31)
	}
	for _, item := range result.Items {
		if city, ok := item.Data.(model.City); ok {
			t.Log(city.Id, city.Name)
		}
	}
}

func TestParseCity(t *testing.T) {
	filename := fmt.Sprintf(localPath, "62.html")
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}
	resp := nami.Response{
		Body: bytes,
		Request: &nami.Request{
			Url: indexUrl + "/62.html",
			Ctx: nami.NewContext(),
		},
	}
	result := ParseCity(resp)
	for _, item := range result.Items {
		log.Printf("%+v\n", item)
	}
}

func TestParseCountyRe(t *testing.T) {
	filename := fmt.Sprintf(localPath, "6206.html")
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}
	resp := nami.Response{
		Body: bytes,
		Request: &nami.Request{
			Url: indexUrl + "/62/6206.html",
			Ctx: nami.NewContext(),
		},
	}
	result := ParseCounty(resp)
	for _, item := range result.Items {
		log.Printf("%+v\n", item)
	}
}

func TestParseTown(t *testing.T) {
	filename := fmt.Sprintf(localPath, "620602.html")
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}
	resp := nami.Response{
		Body: bytes,
		Request: &nami.Request{
			Url: indexUrl + "/62/06/620602.html",
			Ctx: nami.NewContext(),
		},
	}
	result := ParseTown(resp)
	for _, item := range result.Items {
		log.Printf("%+v\n", item)
	}
}

func TestParseVillage(t *testing.T) {
	filename := fmt.Sprintf(localPath, "620602127.html")
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}
	resp := nami.Response{
		Body: bytes,
		Request: &nami.Request{
			Url: indexUrl + "/62/06/02/620602127.html",
			Ctx: nami.NewContext(),
		},
	}
	result := ParseVillage(resp)
	for _, item := range result.Items {
		log.Printf("%+v\n", item)
	}
}
