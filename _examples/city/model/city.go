package model

import "encoding/json"

type City struct {
	ParentId int
	Id       int    //统计用区划代码
	Name     string //名称
	SubId    int    //城乡分类代码
}

func FromJsonObj(o interface{}) (City, error) {
	var profile City
	s, err := json.Marshal(o)
	if err != nil {
		return profile, err
	}
	err = json.Unmarshal(s, &profile)
	return profile, err
}
