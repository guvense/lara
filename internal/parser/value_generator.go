package parser

import (
	"encoding/json"
	"log"
	"strconv"
	"strings"

	"github.com/tidwall/gjson"

	generator "github.com/guvense/lara/internal/generator"
	model "github.com/guvense/lara/internal/model"
)

type Cached struct {
	Rest model.Rest `json:"rest"`
}

func PrepareString(fullText string, extracteds []string, cached *Parser) string {

	prepared := fullText
	for _, value := range extracteds {
		command := strings.Replace(value, "${", "", -1)
		command = strings.Replace(command, "}", "", -1)
		s := strings.Split(command, "::")

		if s[0] == "value" {
			data, _ := json.Marshal(cached)
			test := string(data)
			log.Print(test)
			searchTerm := prepareSearchString(s[1:])
    		extracted := gjson.GetBytes(data, searchTerm).String()
			if extracted != "" {
				prepared = strings.Replace(prepared, value, extracted, -1)
			}
			continue
		} else if s[0] == "number" {
			generatedValue := callDynamically(s[0], s[1], s[2:]...)
			if generatedValue != "" {
				to := "\"" + value + "\""
				prepared = strings.Replace(prepared, to, generatedValue, -1)
			}

		} else {
			generatedValue := callDynamically(s[0], s[1], s[2:]...)
			if generatedValue != "" {
				prepared = strings.Replace(prepared, value, generatedValue, -1)
			}

		}
	}
	return prepared
}

func prepareSearchString(searchTemrs []string) string {

	search := searchTemrs[0]
	for i := 1; i < len(searchTemrs); i++ {
		search = search + "." + searchTemrs[i]
	}
	return search
}


var strFunctionMap = map[string]interface{}{
	"random": random,
	"uuid":   uuidGenerator,
}

var numberFunctionMap = map[string]interface{}{
	"generate": numberGenerator,
	
}

func callDynamically(pack string, name string, args ...string) string {
	switch pack {
	case "str":
		if len(args) == 1 {
			return (strFunctionMap[name].(func(string) string)(args[0]))
		}
		return (strFunctionMap[name].(func() string)())
	case "number": 
		if len(args) == 2 {
			first, err := strconv.ParseInt(args[0], 0, 64)
			if err != nil {
				return ""
			}

			second, err := strconv.ParseInt(args[1], 0, 64)

			if err != nil {
				return ""
			}

			return (numberFunctionMap[name].(func(int64, int64) string)(int64(first), int64(second)))
		}
	}
	return ""
}

func random(value string) string {
	intVar, err := strconv.Atoi(value)
	if err != nil {
		return ""
	}
	return generator.RandomString(intVar)
}

func uuidGenerator() string {

	return generator.GenerateUuid()
}

func numberGenerator(from int64, to int64) string {

	value :=  generator.GenerateNumber(from, to)
	return  strconv.FormatInt(int64(value), 10)
}
