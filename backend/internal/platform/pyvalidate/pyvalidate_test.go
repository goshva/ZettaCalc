package pyvalidate_test

import (
	"testing"

	"github.com/zetta/zettacalc/backend/internal/platform/pyjson"
	"github.com/zetta/zettacalc/backend/internal/platform/pyvalidate"
)

// Эталоны в тестах взяты из ответов Python-mock (vue-prototypes/api/main.py),
// поэтому сравнение идёт по байтам: тексты ошибок показываются пользователю.
func TestModelStrErrors(t *testing.T) {
	cases := []struct {
		name  string
		body  string
		field pyvalidate.StrField
		want  string
	}{
		{
			name:  "короткая строка",
			body:  `{"name":""}`,
			field: pyvalidate.StrField{Name: "name", Required: true, MinLength: 1, MaxLength: 80},
			want: `{"detail":[{"type":"string_too_short","loc":["body","name"],` +
				`"msg":"String should have at least 1 character","input":"",` +
				`"ctx":{"min_length":1}}]}`,
		},
		{
			name:  "длинная строка",
			body:  `{"name":"12345"}`,
			field: pyvalidate.StrField{Name: "name", Required: true, MinLength: 1, MaxLength: 3},
			want: `{"detail":[{"type":"string_too_long","loc":["body","name"],` +
				`"msg":"String should have at most 3 characters","input":"12345",` +
				`"ctx":{"max_length":3}}]}`,
		},
		{
			name:  "не строка",
			body:  `{"name":123}`,
			field: pyvalidate.StrField{Name: "name", Required: true, MinLength: 1},
			want: `{"detail":[{"type":"string_type","loc":["body","name"],` +
				`"msg":"Input should be a valid string","input":123}]}`,
		},
		{
			name:  "отсутствует",
			body:  `{"other":1}`,
			field: pyvalidate.StrField{Name: "name", Required: true, MinLength: 1},
			want: `{"detail":[{"type":"missing","loc":["body","name"],` +
				`"msg":"Field required","input":{"other":1}}]}`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			value, err := pyjson.Decode([]byte(tc.body))
			if err != nil {
				t.Fatalf("разбор тела: %v", err)
			}
			model := pyvalidate.NewModel(value, value)
			model.Str(tc.field)
			if got := string(model.Errors().Body()); got != tc.want {
				t.Errorf("тело ошибки не совпадает\nполучено: %s\nожидалось: %s", got, tc.want)
			}
		})
	}
}

func TestModelStrListErrors(t *testing.T) {
	value, err := pyjson.Decode([]byte(`{"ideas":[1,"a","b","c","d","e"]}`))
	if err != nil {
		t.Fatalf("разбор тела: %v", err)
	}
	model := pyvalidate.NewModel(value, value)
	model.StrList(pyvalidate.ListField{Name: "ideas", MaxItems: 5})
	want := `{"detail":[{"type":"too_long","loc":["body","ideas"],` +
		`"msg":"List should have at most 5 items after validation, not 6",` +
		`"input":[1,"a","b","c","d","e"],` +
		`"ctx":{"field_type":"List","max_length":5,"actual_length":6}}]}`
	if got := string(model.Errors().Body()); got != want {
		t.Errorf("тело ошибки не совпадает\nполучено: %s\nожидалось: %s", got, want)
	}
}

func TestModelBoolLax(t *testing.T) {
        cases := []struct {
                body string
                want bool
        }{
                {`{"consent":1}`, true},
                {`{"consent":0}`, false},
                {`{"consent":"yes"}`, true},
                {`{"consent":"OFF"}`, false},
                {`{"consent":1.0}`, true},
                {`{"consent":-0.0}`, false},
                // Дробные значения, которые pydantic приводит к int.
                {`{"consent":1e-400}`, false},
                {`{"consent":-1e-400}`, false},
                {`{"consent":-0}`, false},
        }
	for _, tc := range cases {
		value, err := pyjson.Decode([]byte(tc.body))
		if err != nil {
			t.Fatalf("разбор тела %s: %v", tc.body, err)
		}
		model := pyvalidate.NewModel(value, value)
		got, ok := model.Bool(pyvalidate.BoolField{Name: "consent"})
		if !ok || got != tc.want {
			t.Errorf("%s: получено %v (ok=%v), ожидалось %v", tc.body, got, ok, tc.want)
		}
		if model.Errors().Len() != 0 {
			t.Errorf("%s: неожиданные ошибки %s", tc.body, model.Errors().Body())
		}
	}
}

func TestModelBoolErrors(t *testing.T) {
	cases := []struct {
		body string
		want string
	}{
		{
			body: `{"consent":"abc"}`,
			want: `{"detail":[{"type":"bool_parsing","loc":["body","consent"],` +
				`"msg":"Input should be a valid boolean, unable to interpret input","input":"abc"}]}`,
		},
		{
                        body: `{"consent":null}`,
                        want: `{"detail":[{"type":"bool_type","loc":["body","consent"],` +
                                `"msg":"Input should be a valid boolean","input":null}]}`,
                },
                // Дробное число и число вне int64 pydantic не приводит к int,
                // поэтому это bool_type, а не bool_parsing.
                {
                        body: `{"consent":1.5}`,
                        want: `{"detail":[{"type":"bool_type","loc":["body","consent"],` +
                                `"msg":"Input should be a valid boolean","input":1.5}]}`,
                },
                {
                        body: `{"consent":0.5}`,
                        want: `{"detail":[{"type":"bool_type","loc":["body","consent"],` +
                                `"msg":"Input should be a valid boolean","input":0.5}]}`,
                },
                {
                        body: `{"consent":9223372036854775808}`,
                        want: `{"detail":[{"type":"bool_type","loc":["body","consent"],` +
                                `"msg":"Input should be a valid boolean","input":9223372036854775808}]}`,
                },
                {
                        body: `{"consent":2}`,
                        want: `{"detail":[{"type":"bool_parsing","loc":["body","consent"],` +
                                `"msg":"Input should be a valid boolean, unable to interpret input","input":2}]}`,
                },
        }
	for _, tc := range cases {
		value, err := pyjson.Decode([]byte(tc.body))
		if err != nil {
			t.Fatalf("разбор тела %s: %v", tc.body, err)
		}
		model := pyvalidate.NewModel(value, value)
		model.Bool(pyvalidate.BoolField{Name: "consent"})
		if got := string(model.Errors().Body()); got != tc.want {
			t.Errorf("тело ошибки не совпадает\nполучено: %s\nожидалось: %s", got, tc.want)
		}
	}
}

// Даты проверяются отдельно: тексты ошибок speedate длинные и дословные.
func TestDateErrors(t *testing.T) {
	cases := []struct {
		value string
		want  string
	}{
		{
			value: `"zz"`,
			want: `{"detail":[{"type":"date_from_datetime_parsing","loc":["body","preferred_date"],` +
				`"msg":"Input should be a valid date or datetime, input is too short","input":"zz",` +
				`"ctx":{"error":"input is too short"}}]}`,
		},
		{
			value: `"2026-13-01"`,
			want: `{"detail":[{"type":"date_from_datetime_parsing","loc":["body","preferred_date"],` +
				`"msg":"Input should be a valid date or datetime, month value is outside expected range of 1-12",` +
				`"input":"2026-13-01","ctx":{"error":"month value is outside expected range of 1-12"}}]}`,
		},
		{
			value: `"2026-11-06T25:00:00"`,
			want: `{"detail":[{"type":"date_from_datetime_parsing","loc":["body","preferred_date"],` +
				`"msg":"Input should be a valid date or datetime, hour value is outside expected range of 0-23",` +
				`"input":"2026-11-06T25:00:00",` +
				`"ctx":{"error":"hour value is outside expected range of 0-23"}}]}`,
		},
		{
			value: `"2026-11-06 12:00:00"`,
			want: `{"detail":[{"type":"date_from_datetime_inexact","loc":["body","preferred_date"],` +
				`"msg":"Datetimes provided to dates should have zero time - e.g. be exact dates",` +
				`"input":"2026-11-06 12:00:00"}]}`,
		},
		{
			value: `123`,
			want: `{"detail":[{"type":"date_from_datetime_inexact","loc":["body","preferred_date"],` +
				`"msg":"Datetimes provided to dates should have zero time - e.g. be exact dates",` +
				`"input":123}]}`,
		},
		{
			value: `true`,
			want: `{"detail":[{"type":"date_type","loc":["body","preferred_date"],` +
				`"msg":"Input should be a valid date","input":true}]}`,
		},
		{
			value: `"0000-01-01"`,
			want: `{"detail":[{"type":"date_parsing","loc":["body","preferred_date"],` +
				`"msg":"Input should be a valid date in the format YYYY-MM-DD, year 0 is out of range",` +
				`"input":"0000-01-01","ctx":{"error":"year 0 is out of range"}}]}`,
		},
	}

	for _, tc := range cases {
		body := `{"preferred_date":` + tc.value + `}`
		value, err := pyjson.Decode([]byte(body))
		if err != nil {
			t.Fatalf("разбор тела %s: %v", body, err)
		}
		model := pyvalidate.NewModel(value, value)
		model.Date("preferred_date")
		if got := string(model.Errors().Body()); got != tc.want {
			t.Errorf("%s\nполучено: %s\nожидалось: %s", tc.value, got, tc.want)
		}
	}
}

// Допустимые даты: mock API нормализует их к YYYY-MM-DD.
func TestDateAccepted(t *testing.T) {
	cases := map[string]string{
		`"2026-11-06"`:                  "2026-11-06",
		`"2026-11-06T00:00:00Z"`:        "2026-11-06",
		`"2026-11-06t00:00:00"`:         "2026-11-06",
		`"2026-11-06_00:00:00"`:         "2026-11-06",
		`"2026-11-06 00:00"`:            "2026-11-06",
		`"2026-11-06T00:00:00+03:00"`:   "2026-11-06",
		`"2026-11-06T00:00:00.0000001"`: "2026-11-06",
		`"2024-02-29"`:                  "2024-02-29",
		`0`:                             "1970-01-01",
		`86400`:                         "1970-01-02",
		`"0"`:                           "1970-01-01",
		`null`:                          "",
	}

	for body, want := range cases {
		value, err := pyjson.Decode([]byte(`{"preferred_date":` + body + `}`))
		if err != nil {
			t.Fatalf("разбор тела %s: %v", body, err)
		}
		model := pyvalidate.NewModel(value, value)
		got, ok := model.Date("preferred_date")
		if !ok {
			t.Errorf("%s: неожиданная ошибка %s", body, model.Errors().Body())
			continue
		}
		if got != want {
			t.Errorf("%s: получено %q, ожидалось %q", body, got, want)
		}
	}
}
