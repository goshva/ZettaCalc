package pyvalidate

import (
	"math"
	"strings"

	"github.com/zetta/zettacalc/backend/internal/platform/pyjson"
)

// StrField описывает поле `str` в модели pydantic.
type StrField struct {
	// Name — имя поля в loc.
	Name string
	// Required — поле обязательно (без default).
	Required bool
	// MinLength — минимальная длина (0 — без ограничения).
	MinLength int
	// MaxLength — максимальная длина (0 — без ограничения).
	MaxLength int
	// Default — значение по умолчанию для отсутствующего поля.
	Default string
}

// ListField описывает поле `list[str]` в модели pydantic.
type ListField struct {
	Name     string
	Required bool
	MaxItems int
}

// BoolField описывает поле `bool`.
type BoolField struct {
	Name     string
	Required bool
	Default  bool
}

// Model — набор полей в порядке объявления, как в pydantic-модели.
type Model struct {
	// Object — разобранный объект тела.
	Object *pyjson.Value
	// Raw — исходное значение для loc ["body"] в ошибках missing.
	Raw *pyjson.Value

	errs *Errors
}

// NewModel создаёт модель поверх разобранного объекта.
func NewModel(object, raw *pyjson.Value) *Model {
	return &Model{Object: object, Raw: raw, errs: &Errors{}}
}

// Errors возвращает накопленные ошибки.
func (m *Model) Errors() *Errors { return m.errs }

// Value возвращает значение поля или nil.
func (m *Model) Value(name string) *pyjson.Value {
	if m.Object == nil {
		return nil
	}
	v, _ := m.Object.Get(name)
	return v
}

// Has сообщает, что поле присутствует в теле.
func (m *Model) Has(name string) bool {
	if m.Object == nil {
		return false
	}
	_, ok := m.Object.Get(name)
	return ok
}

// Str разбирает поле `str`.
//
// Возвращает значение, признак допустимости и признак наличия ошибки. Отдельно
// сообщается «поле отсутствовало», чтобы вызывающий код мог отличить default
// от явного значения.
func (m *Model) Str(f StrField) (string, bool) {
	v := m.Value(f.Name)
	if v == nil {
		if f.Required {
			m.errs.Add(Error{
				Type:  "missing",
				Loc:   Body(f.Name),
				Msg:   "Field required",
				Input: m.Raw,
			})
			return "", false
		}
		return f.Default, true
	}
	if v.Kind == pyjson.Null {
		// null не подставляется вместо default: pydantic сообщает string_type.
		m.errs.Add(Error{
			Type:  "string_type",
			Loc:   Body(f.Name),
			Msg:   "Input should be a valid string",
			Input: v,
		})
		return "", false
	}
	if v.Kind != pyjson.String {
		m.errs.Add(Error{
			Type:  "string_type",
			Loc:   Body(f.Name),
			Msg:   "Input should be a valid string",
			Input: v,
		})
		return "", false
	}
	length := pyjson.RuneLen(v.Str)
	if f.MinLength > 0 && length < f.MinLength {
		m.errs.Add(Error{
			Type: "string_too_short",
			Loc:  Body(f.Name),
			Msg: "String should have at least " + itoa(f.MinLength) + " " +
				PluralCharacters(f.MinLength),
			Input: v,
			Ctx:   NewCtx().SetInt("min_length", f.MinLength),
		})
		return "", false
	}
	if f.MaxLength > 0 && length > f.MaxLength {
		m.errs.Add(Error{
			Type: "string_too_long",
			Loc:  Body(f.Name),
			Msg: "String should have at most " + itoa(f.MaxLength) + " " +
				PluralCharacters(f.MaxLength),
			Input: v,
			Ctx:   NewCtx().SetInt("max_length", f.MaxLength),
		})
		return "", false
	}
	return v.Str, true
}

// StrList разбирает поле `list[str]`.
func (m *Model) StrList(f ListField) ([]string, bool) {
	v := m.Value(f.Name)
	if v == nil {
		if f.Required {
			m.errs.Add(Error{
				Type:  "missing",
				Loc:   Body(f.Name),
				Msg:   "Field required",
				Input: m.Raw,
			})
			return nil, false
		}
		return []string{}, true
	}
	if v.Kind == pyjson.Null {
		m.errs.Add(Error{
			Type:  "list_type",
			Loc:   Body(f.Name),
			Msg:   "Input should be a valid list",
			Input: v,
		})
		return nil, false
	}
	if v.Kind != pyjson.Array {
		m.errs.Add(Error{
			Type:  "list_type",
			Loc:   Body(f.Name),
			Msg:   "Input should be a valid list",
			Input: v,
		})
		return nil, false
	}

	// pydantic проверяет длину списка раньше элементов.
	if f.MaxItems > 0 && len(v.Items) > f.MaxItems {
		m.errs.Add(Error{
			Type: "too_long",
			Loc:  Body(f.Name),
			Msg: "List should have at most " + itoa(f.MaxItems) + " " +
				PluralItems(f.MaxItems) + " after validation, not " + itoa(len(v.Items)),
			Input: v,
			Ctx: NewCtx().
				SetString("field_type", "List").
				SetInt("max_length", f.MaxItems).
				SetInt("actual_length", len(v.Items)),
		})
		return nil, false
	}

	out := make([]string, 0, len(v.Items))
	ok := true
	for i, item := range v.Items {
		if item.Kind != pyjson.String {
			m.errs.Add(Error{
				Type:  "string_type",
				Loc:   Body(f.Name, i),
				Msg:   "Input should be a valid string",
				Input: item,
			})
			ok = false
			continue
		}
		out = append(out, item.Str)
	}
	if !ok {
		return nil, false
	}
	return out, true
}

// Bool разбирает поле `bool` в lax-режиме pydantic.
func (m *Model) Bool(f BoolField) (bool, bool) {
	v := m.Value(f.Name)
	if v == nil {
		if f.Required {
			m.errs.Add(Error{
				Type:  "missing",
				Loc:   Body(f.Name),
				Msg:   "Field required",
				Input: m.Raw,
			})
			return false, false
		}
		return f.Default, true
	}
	switch v.Kind {
	case pyjson.Null:
		m.errs.Add(Error{
			Type:  "bool_type",
			Loc:   Body(f.Name),
			Msg:   "Input should be a valid boolean",
			Input: v,
		})
		return false, false
	case pyjson.Bool:
		return v.BoolValue, true
	case pyjson.Int:
		if v.IntValue.Sign() == 0 {
			return false, true
		}
		// pydantic сначала приводит значение к int, а числа вне int64
		// не приводятся — это уже не «неинтерпретируемое значение», а тип.
		if !v.IntValue.IsInt64() {
			return m.boolTypeError(f.Name, v)
		}
		if v.IntValue.Int64() == 1 {
			return true, true
		}
		return m.boolParsingError(f.Name, v)
	case pyjson.Float:
		if math.IsNaN(v.FloatVal) || math.IsInf(v.FloatVal, 0) {
			return m.boolTypeError(f.Name, v)
		}
		if v.FloatVal != math.Trunc(v.FloatVal) {
			return m.boolTypeError(f.Name, v)
		}
		if v.FloatVal < math.MinInt64 || v.FloatVal >= 9223372036854775808.0 {
			return m.boolTypeError(f.Name, v)
		}
		if v.FloatVal == 0 {
			return false, true
		}
		if v.FloatVal == 1 {
			return true, true
		}
		return m.boolParsingError(f.Name, v)
	case pyjson.String:
		switch strings.ToLower(v.Str) {
		case "true", "yes", "on", "1", "t", "y":
			return true, true
		case "false", "no", "off", "0", "f", "n":
			return false, true
		}
		return m.boolParsingError(f.Name, v)
	default:
		m.errs.Add(Error{
			Type:  "bool_type",
			Loc:   Body(f.Name),
			Msg:   "Input should be a valid boolean",
			Input: v,
		})
		return false, false
	}
}

func (m *Model) boolParsingError(name string, v *pyjson.Value) (bool, bool) {
	m.errs.Add(Error{
		Type:  "bool_parsing",
		Loc:   Body(name),
		Msg:   "Input should be a valid boolean, unable to interpret input",
		Input: v,
	})
	return false, false
}

func (m *Model) boolTypeError(name string, v *pyjson.Value) (bool, bool) {
	m.errs.Add(Error{
		Type:  "bool_type",
		Loc:   Body(name),
		Msg:   "Input should be a valid boolean",
		Input: v,
	})
	return false, false
}

// Date разбирает поле `date | None`.
//
// Отсутствие поля и null допустимы: значение просто не заполнено.
func (m *Model) Date(name string) (string, bool) {
	v := m.Value(name)
	if v == nil {
		return "", true
	}
	if v.Kind == pyjson.Null {
		return "", true
	}
	result := DateValue(v)
	if result.Err != nil {
		err := *result.Err
		err.Loc = Body(name)
		m.errs.Add(err)
		return "", false
	}
	return result.Value, true
}

// MissingIfNil добавляет ошибку missing для обязательного поля, которого нет.
func (m *Model) MissingIfNil(name string) {
	if !m.Has(name) {
		m.errs.Add(Error{
			Type:  "missing",
			Loc:   Body(name),
			Msg:   "Field required",
			Input: m.Raw,
		})
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
