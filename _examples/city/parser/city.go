package parser

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/opreader/nami"
	"github.com/opreader/nami/_examples/city/model"
)

const (
	localPath = `/Users/luffy/dev/nami/_examples/city/files/%s`
	indexUrl  = `http://www.stats.gov.cn/tjsj/tjbz/tjyqhdmhcxhfdm/2020`
)

var (
	provinceRe = regexp.MustCompile(`<td><a href='(\d+\.html)'>(.+?)<br/></a></td>`)

	cityRe = regexp.MustCompile(`<tr class='citytr'><td><a href='([\d/]+.html)'>(\d+)</a></td><td><a href='([\d/]+.html)'>(.+?)</a></td></tr>`)

	countyRe = regexp.MustCompile(`<tr class='countytr'><td><a href='([\d/]+.html)'>(\d+)</a></td><td><a href='([\d/]+.html)'>(.+?)</a></td></tr>`)

	townRe = regexp.MustCompile(`<tr class='towntr'><td><a href='([\d/]+.html)'>(\d+)</a></td><td><a href='([\d/]+.html)'>(.+?)</a></td></tr>`)

	villageRe = regexp.MustCompile(`<tr class='villagetr'><td>(\d+)</td><td>(\d+)</td><td>(.+?)</td></tr>`)

	districtRe = regexp.MustCompile(`<tr class='countytr'><td>(\d+)</td><td>市辖区</td></tr>`)
)

func ParseProvince(resp nami.Response) nami.Result {
	//go createFile(resp.Body, "country.html")
	result := nami.Result{}
	matches := provinceRe.FindAllSubmatch(resp.Body, -1)
	for _, m := range matches {
		url := fmt.Sprintf("%s/%s", indexUrl, m[1])
		idStr := string(m[1])
		id, _ := strconv.Atoi(idStr[:len(idStr)-5])
		name := string(m[2])
		item := nami.Item{
			Url:  url,
			Data: model.City{Id: id, Name: name},
		}
		result.Items = append(result.Items, item)
		ctxMap := map[string]interface{}{"parentId": id}
		result.AddTask(url, ParseCity, ctxMap)
	}
	return result
}

func ParseCity(resp nami.Response) nami.Result {
	/*	if resp.Request.Url == indexUrl+"/62.html" {
		go createFile(resp.Body, "62.html")
	}*/
	parentId := 0
	if parentIdValue := resp.Request.Ctx.GetAny("parentId"); parentIdValue != nil {
		parentId = parentIdValue.(int)
	}
	result := nami.Result{}
	matches := cityRe.FindAllStringSubmatch(string(resp.Body), -1)
	for _, m := range matches {
		url := fmt.Sprintf("%s/%s", indexUrl, m[1])
		id, _ := strconv.Atoi(m[2])
		name := m[4]
		item := nami.Item{
			Url:  url,
			Data: model.City{ParentId: parentId, Id: id, Name: name},
		}
		result.Items = append(result.Items, item)
		ctxMap := map[string]interface{}{"parentId": id}
		result.AddTask(url, ParseCounty, ctxMap)
	}
	return result
}

func ParseCounty(resp nami.Response) nami.Result {
	/*	if resp.Request.Url == indexUrl+"/62/6206.html" {
		go createFile(resp.Body, "6206.html")
	}*/
	parentId := 0
	if parentIdValue := resp.Request.Ctx.GetAny("parentId"); parentIdValue != nil {
		parentId = parentIdValue.(int)
	}
	result := nami.Result{}
	if item, ok := parseDistrict(parentId, string(resp.Body)); ok {
		result.Items = append(result.Items, item)
	}
	index := strings.LastIndex(resp.Request.Url, "/")
	prefix := resp.Request.Url[0:index]
	matches := countyRe.FindAllStringSubmatch(string(resp.Body), -1)
	for _, m := range matches {
		url := fmt.Sprintf("%s/%s", prefix, m[1])
		id, _ := strconv.Atoi(m[2])
		name := m[4]
		item := nami.Item{
			Url:  url,
			Data: model.City{ParentId: parentId, Id: id, Name: name},
		}
		result.Items = append(result.Items, item)
		ctxMap := map[string]interface{}{"parentId": id}
		result.AddTask(url, ParseTown, ctxMap)
	}
	return result
}

func ParseTown(resp nami.Response) nami.Result {
	/*	if resp.Request.Url == indexUrl+"/62/06/620602.html" {
		go createFile(resp.Body, "620602.html")
	}*/
	parentId := 0
	if parentIdValue := resp.Request.Ctx.GetAny("parentId"); parentIdValue != nil {
		parentId = parentIdValue.(int)
	}
	result := nami.Result{}
	index := strings.LastIndex(resp.Request.Url, "/")
	prefix := resp.Request.Url[0:index]
	matches := townRe.FindAllStringSubmatch(string(resp.Body), -1)
	for _, m := range matches {
		url := fmt.Sprintf("%s/%s", prefix, m[1])
		id, _ := strconv.Atoi(m[2])
		name := m[4]
		item := nami.Item{
			Url:  url,
			Data: model.City{ParentId: parentId, Id: id, Name: name},
		}
		result.Items = append(result.Items, item)
		ctxMap := map[string]interface{}{"parentId": id}
		result.AddTask(url, ParseVillage, ctxMap)
	}
	return result
}

func ParseVillage(resp nami.Response) nami.Result {
	/*	if resp.Request.Url == indexUrl+"/62/06/02/620602127.html" {
		go createFile(resp.Body, "620602127.html")
	}*/
	parentId := 0
	if parentIdValue := resp.Request.Ctx.GetAny("parentId"); parentIdValue != nil {
		parentId = parentIdValue.(int)
	}
	result := nami.Result{}
	matches := villageRe.FindAllStringSubmatch(string(resp.Body), -1)
	for _, m := range matches {
		id, _ := strconv.Atoi(m[1])
		subId, _ := strconv.Atoi(m[2])
		item := nami.Item{
			Url: resp.Request.Url,
			Data: model.City{
				ParentId: parentId,
				Id:       id,
				SubId:    subId,
				Name:     m[3],
			},
		}
		result.Items = append(result.Items, item)
	}
	return result
}

func parseDistrict(parentId int, s string) (nami.Item, bool) {
	item := nami.Item{}
	city := model.City{ParentId: parentId, Name: "市辖区"}
	match := districtRe.FindStringSubmatch(s)
	if match != nil {
		id, _ := strconv.Atoi(match[1])
		city.Id = id
		item.Data = city
		return item, true
	}
	return item, false
}

func createFile(bytes []byte, name string) {
	filename := fmt.Sprintf(localPath, name)
	err := ioutil.WriteFile(filename, bytes, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
}

func bytes2int(buf []byte) int {
	var data int
	buffer := bytes.NewBuffer(buf)
	binary.Read(buffer, binary.BigEndian, &data)
	return data
}
