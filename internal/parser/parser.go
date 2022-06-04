package parser

import (
	"encoding/json"
	"regexp"

	lara "github.com/guvense/lara/internal"
)

type Parser struct {

	Request RequestParser `json:"request"`
	Response ResponseParser `json:"response"`
	Config  lara.Config 
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
