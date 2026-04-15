package parser

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	lara "github.com/guvense/lara/internal"
)

func TestParseReplacesValueAndUuid(t *testing.T) {
	p := Parser{
		Request: RequestParser{
			Params: &map[string]string{"username": "john"},
			Body:   json.RawMessage(`{"name":"alice"}`),
		},
	}

	parsed := p.Parse(`{"u":"${value::request::queryparams::username}","id":"${str::uuid}"}`)

	if !strings.Contains(parsed, `"u":"john"`) {
		t.Fatalf("expected query param replacement, got %s", parsed)
	}
	if strings.Contains(parsed, "${str::uuid}") {
		t.Fatalf("expected uuid placeholder to be replaced, got %s", parsed)
	}
}

func TestPrepareStringNumberGenerationUnquotesNumericValue(t *testing.T) {
	result := PrepareString(`{"point":"${number::generate::1::10}"}`, []string{"${number::generate::1::10}"}, &Parser{})

	if strings.Contains(result, `"${number::generate::1::10}"`) {
		t.Fatalf("expected number placeholder to be replaced, got %s", result)
	}
	if strings.Contains(result, `"point":"`) {
		t.Fatalf("expected numeric value to stay unquoted, got %s", result)
	}
}

func TestPrepareStringDateAndRegexFromConfig(t *testing.T) {
	p := &Parser{
		Config: lara.Config{
			RegexExpression: map[string]string{"ten": "^\\d{10}$"},
			Date:            map[string]string{"iso": "2006-01-02"},
		},
	}

	input := `{"date":"${date::now::iso}","code":"${str::regex::ten}"}`
	out := p.Parse(input)

	if strings.Contains(out, "${date::now::iso}") || strings.Contains(out, "${str::regex::ten}") {
		t.Fatalf("expected placeholders to be replaced, got %s", out)
	}

	var body map[string]string
	if err := json.Unmarshal([]byte(out), &body); err != nil {
		t.Fatalf("failed to unmarshal output: %v, output: %s", err, out)
	}

	if _, err := time.Parse("2006-01-02", body["date"]); err != nil {
		t.Fatalf("expected iso date, got %q", body["date"])
	}
	if len(body["code"]) != 10 {
		t.Fatalf("expected 10-digit code, got %q", body["code"])
	}
}

func TestCallDynamicallyInvalidNumberInputReturnsEmpty(t *testing.T) {
	if got := callDynamically("number", "generate", "a", "2"); got != "" {
		t.Fatalf("expected empty result for invalid input, got %q", got)
	}
}

func TestPrepareSearchString(t *testing.T) {
	got := prepareSearchString([]string{"request", "body", "username"})
	if got != "request.body.username" {
		t.Fatalf("expected dot-separated search term, got %s", got)
	}
}
