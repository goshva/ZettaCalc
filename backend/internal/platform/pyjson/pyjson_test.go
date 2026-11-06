package pyjson_test

import (
	"math"
	"testing"

	"github.com/zetta/zettacalc/backend/internal/platform/pyjson"
)

// Эталоны взяты из ответов json.loads в CPython: non-finite литералы
// принимаются при разборе, хотя стандарт JSON их не разрешает.
func TestDecodeNonFinite(t *testing.T) {
	inf := math.Inf(1)
	negInf := math.Inf(-1)
	nan := math.NaN()

	cases := []struct {
		body    string
		want    *float64
		invalid bool
	}{
		{body: `Infinity`, want: &inf},
		{body: `-Infinity`, want: &negInf},
		{body: `NaN`, want: &nan},
		{body: `{"x":Infinity}`, want: &inf},
		{body: `{"x":-Infinity}`, want: &negInf},
		{body: `[NaN]`, want: &nan},
		{body: `1e400`, want: &inf},
		{body: `-1e400`, want: &negInf},
		{body: `{"x":1e400}`, want: &inf},
		// Строчные написания json.loads не принимает.
		{body: `nan`, invalid: true},
		{body: `inf`, invalid: true},
		{body: `infinity`, invalid: true},
		{body: `-Infinit`, invalid: true},
	}

	for _, tc := range cases {
		value, err := pyjson.Decode([]byte(tc.body))
		if tc.invalid {
			if err == nil {
				t.Errorf("%s: ожидалась ошибка разбора, получено %v", tc.body, value)
			}
			continue
		}
		if err != nil {
			t.Errorf("%s: неожиданная ошибка разбора: %v", tc.body, err)
			continue
		}
		if !pyjson.HasNonFinite(value) {
			t.Errorf("%s: значение не распознано как неконечное", tc.body)
			continue
		}
		leaf := value
		switch value.Kind {
		case pyjson.Object:
			leaf = value.Values[0]
		case pyjson.Array:
			leaf = value.Items[0]
		}
		if math.IsNaN(*tc.want) {
			if !math.IsNaN(leaf.FloatVal) {
				t.Errorf("%s: получено %v, ожидался NaN", tc.body, leaf.FloatVal)
			}
			continue
		}
		if leaf.FloatVal != *tc.want {
			t.Errorf("%s: получено %v, ожидалось %v", tc.body, leaf.FloatVal, *tc.want)
		}
	}
}

// Переполнение мантиссы без литерала Infinity: 1e-400 даёт обычный ноль.
func TestDecodeUnderflow(t *testing.T) {
	value, err := pyjson.Decode([]byte(`1e-400`))
	if err != nil {
		t.Fatalf("разбор: %v", err)
	}
	if value.Kind != pyjson.Float || value.FloatVal != 0 {
		t.Errorf("получено %v, ожидался 0.0", value)
	}
	if pyjson.HasNonFinite(value) {
		t.Error("0.0 распознан как неконечное значение")
	}
}

func TestDecodeNonFiniteTail(t *testing.T) {
	// Хвост после литерала — обычная синтаксическая ошибка, как в json.loads.
	cases := []struct {
		body string
		pos  int
	}{
		{body: `Infinity1`, pos: 8},
		{body: `{"x":Infinity1}`, pos: 13},
		{body: `NaNx`, pos: 3},
		{body: `-Infinityx`, pos: 9},
	}
	for _, tc := range cases {
		_, err := pyjson.Decode([]byte(tc.body))
		syntax, ok := err.(*pyjson.SyntaxError)
		if !ok {
			t.Errorf("%s: ожидалась синтаксическая ошибка, получено %v", tc.body, err)
			continue
		}
		if syntax.Pos != tc.pos {
			t.Errorf("%s: позиция %d, ожидалось %d", tc.body, syntax.Pos, tc.pos)
		}
	}
}

func TestHasNonFinite(t *testing.T) {
	value, err := pyjson.Decode([]byte(`{"a":[1,{"b":"x"}]}`))
	if err != nil {
		t.Fatalf("разбор: %v", err)
	}
	if pyjson.HasNonFinite(value) {
		t.Error("обычное значение распознано как неконечное")
	}
}

func TestAsIntRejectsNonFinite(t *testing.T) {
	// int(inf) и int(nan) в Python — ошибка, поэтому mock API отвечает 500.
	for _, body := range []string{`Infinity`, `-Infinity`, `NaN`} {
		value, err := pyjson.Decode([]byte(body))
		if err != nil {
			t.Fatalf("разбор %s: %v", body, err)
		}
		if _, err := pyjson.AsInt(value); err == nil {
			t.Errorf("%s: int() неожиданно принял неконечное значение", body)
		}
	}
}
