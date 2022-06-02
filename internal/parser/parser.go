package parser

import (
	"encoding/json"
	"regexp"
)

type Parser struct {

	Request RequestParser `json:"request"`
	Response ResponseParser `json:"response"`
}

type RequestParser struct {
	Params   *map[string]string `json:"queryparams"`
	PathVariables   *map[string]string `json:"pathvariables"`
	Body json.RawMessage `json:"body"`
}

type ResponseParser struct {
	Body json.RawMessage `json:"body"`
}

func (p *Parser) Parse(value string) string {

	regexString := `\${(.*?)\}`

	re := regexp.MustCompile(regexString)

	submatchall := re.FindAllString(value, -1)

	var matcheds []string
	for _, element := range submatchall {
		matcheds = append(matcheds, element)
	}

	return PrepareString(value, matcheds, p)
}
