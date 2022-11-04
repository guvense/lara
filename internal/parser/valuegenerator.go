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

var strFunctionMap = map[string]interface{}{
	"random": random,
	"uuid":   uuidGenerator,
	"regex":  generateStringFromRegexAndLenght,
}

var dateFunctionMap = map[string]interface{}{
	"now":    generateDateNow,
	"random": generateDateRandom,
}

var numberFunctionMap = map[string]interface{}{
	"generate": numberGenerator,
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

		} else if s[0] == "str" {
			if s[1] == "regex" && cached.Config.RegexExpression != nil {
				expression := cached.Config.RegexExpression[s[2]]
				if expression == "" {
					continue
				}
				s[2] = expression

				if len(s) == 3 {
					s = append(s, "0")
				}
			}
			generatedValue := callDynamically(s[0], s[1], s[2:]...)
			if generatedValue != "" {
				prepared = strings.Replace(prepared, value, generatedValue, -1)
			}

		} else if s[0] == "date" {
			dateFormat := "01-02-2006"
			formatIndex := len(s) - 1
			if cached.Config.Date != nil && cached.Config.Date[s[formatIndex]] != "" {
				dateFormat = cached.Config.Date[s[formatIndex]]
			}
			s[formatIndex] = dateFormat
			generatedValue := callDynamically(s[0], s[1], s[2:]...)
			if generatedValue != "" {
				prepared = strings.Replace(prepared, value, generatedValue, -1)
			}
		}
	}
	return prepared
}

func callDynamically(pack string, name string, args ...string) string {
	switch pack {
	case "str":
		if len(args) == 1 {
			return (strFunctionMap[name].(func(string) string)(args[0]))
		} else if len(args) == 2 {

			return (strFunctionMap[name].(func(string, string) string)(args[0], args[1]))
		}
		return (strFunctionMap[name].(func() string)())
	case "date":
		if len(args) == 1 {
			return (dateFunctionMap[name].(func(string) string)(args[0]))
		}
		return ""
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

func generateDateNow(format string) string {
	return generator.GenerateCurrentDayByFormat(format)
}

func generateDateRandom(format string) string {
	return generator.GenerateRandomDateByFormat(format)
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

func generateStringFromRegexAndLenght(regex string, lenght string) string {

	len := 0
	if lenght != "" {
		parsed, err := strconv.Atoi(lenght)
		if err != nil {
			return ""
		}
		len = parsed
	}

	return generator.GenerateStringFromRegexAndLength(regex, len)
}

func numberGenerator(from int64, to int64) string {

	value := generator.GenerateNumber(from, to)
	return strconv.FormatInt(int64(value), 10)
}

func prepareSearchString(searchTemrs []string) string {

	search := searchTemrs[0]
	for i := 1; i < len(searchTemrs); i++ {
		search = search + "." + searchTemrs[i]
	}
	return search
}
