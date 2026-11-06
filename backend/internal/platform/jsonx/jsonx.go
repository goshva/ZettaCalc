// Package jsonx — компактный JSON с сохранением порядка ключей.
//
// Зачем это нужно: паритет с mock API проверяется сравнением тел ответов.
// Python (FastAPI) сериализует так:
//
//	json.dumps(obj, ensure_ascii=False, allow_nan=False, separators=(",", ":"))
//
// то есть порядок ключей сохраняется как в исходном файле, пробелов нет,
// не-ASCII символы не экранируются. Стандартный encoding/json сортирует ключи
// map, поэтому используем собственное упорядоченное дерево.
package jsonx

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Kind — тип значения.
type Kind uint8

const (
	// Null — литерал null.
	Null Kind = iota
	// Bool — true/false.
	Bool
	// Number — число в исходной записи.
	Number
	// String — строка (декодированная).
	String
	// Array — массив.
	Array
	// Object — объект с сохранённым порядком ключей.
	Object
)

// Value — узел упорядоченного JSON.
type Value struct {
	Kind Kind

	Bool   bool
	Number string
	Str    string

	Items  []*Value
	Keys   []string
	Values []*Value
}

// Get возвращает значение поля объекта.
func (v *Value) Get(key string) (*Value, bool) {
	if v == nil || v.Kind != Object {
		return nil, false
	}
	for i, k := range v.Keys {
		if k == key {
			return v.Values[i], true
		}
	}
	return nil, false
}

// Field возвращает значение поля объекта или nil.
//
// В отличие от Get не требует проверки второго значения там, где отсутствие
// поля означает «значение по умолчанию»: узел nil безопасен для всех геттеров.
func (v *Value) Field(key string) *Value {
	item, _ := v.Get(key)
	return item
}

// Set задаёт поле объекта, сохраняя позицию существующего ключа.
func (v *Value) Set(key string, val *Value) {
	if v.Kind != Object {
		return
	}
	for i, k := range v.Keys {
		if k == key {
			v.Values[i] = val
			return
		}
	}
	v.Keys = append(v.Keys, key)
	v.Values = append(v.Values, val)
}

// StringValue — строковое значение узла (пустая строка для не-строк).
func (v *Value) StringValue() string {
	if v == nil || v.Kind != String {
		return ""
	}
	return v.Str
}

// FloatValue — число как float64.
func (v *Value) FloatValue() (float64, bool) {
	if v == nil || v.Kind != Number {
		return 0, false
	}
	f, err := strconv.ParseFloat(v.Number, 64)
	if err != nil {
		return 0, false
	}
	return f, true
}

// BoolValue — булево значение.
func (v *Value) BoolValue() bool { return v != nil && v.Kind == Bool && v.Bool }

// IsNull сообщает, что узел — null или отсутствует.
func (v *Value) IsNull() bool { return v == nil || v.Kind == Null }

// ArrayItems возвращает элементы массива.
func (v *Value) ArrayItems() []*Value {
	if v == nil || v.Kind != Array {
		return nil
	}
	return v.Items
}

// NewObject создаёт пустой объект.
func NewObject() *Value { return &Value{Kind: Object} }

// NewString создаёт строковое значение.
func NewString(s string) *Value { return &Value{Kind: String, Str: s} }

// NewArray создаёт массив.
func NewArray(items ...*Value) *Value { return &Value{Kind: Array, Items: items} }

// Parse разбирает JSON, сохраняя порядок ключей и исходную запись чисел.
func Parse(data []byte) (*Value, error) {
	p := &parser{data: data}
	p.skipSpace()
	v, err := p.value()
	if err != nil {
		return nil, err
	}
	p.skipSpace()
	if p.pos != len(p.data) {
		return nil, fmt.Errorf("jsonx: лишние данные на позиции %d", p.pos)
	}
	return v, nil
}

// ParseString разбирает JSON из строки.
func ParseString(s string) (*Value, error) { return Parse([]byte(s)) }

type parser struct {
	data []byte
	pos  int
}

func (p *parser) skipSpace() {
	for p.pos < len(p.data) {
		switch p.data[p.pos] {
		case ' ', '\t', '\n', '\r':
			p.pos++
		default:
			return
		}
	}
}

var errUnexpectedEnd = errors.New("jsonx: неожиданный конец данных")

func (p *parser) value() (*Value, error) {
	if p.pos >= len(p.data) {
		return nil, errUnexpectedEnd
	}
	switch c := p.data[p.pos]; {
	case c == '{':
		return p.object()
	case c == '[':
		return p.array()
	case c == '"':
		s, err := p.string()
		if err != nil {
			return nil, err
		}
		return &Value{Kind: String, Str: s}, nil
	case c == 't':
		if err := p.literal("true"); err != nil {
			return nil, err
		}
		return &Value{Kind: Bool, Bool: true}, nil
	case c == 'f':
		if err := p.literal("false"); err != nil {
			return nil, err
		}
		return &Value{Kind: Bool, Bool: false}, nil
	case c == 'n':
		if err := p.literal("null"); err != nil {
			return nil, err
		}
		return &Value{Kind: Null}, nil
	default:
		return p.number()
	}
}

func (p *parser) literal(lit string) error {
	if len(p.data)-p.pos < len(lit) || string(p.data[p.pos:p.pos+len(lit)]) != lit {
		return fmt.Errorf("jsonx: ожидался литерал %q на позиции %d", lit, p.pos)
	}
	p.pos += len(lit)
	return nil
}

func (p *parser) number() (*Value, error) {
	start := p.pos
	for p.pos < len(p.data) {
		c := p.data[p.pos]
		if (c >= '0' && c <= '9') || c == '-' || c == '+' || c == '.' || c == 'e' || c == 'E' {
			p.pos++
			continue
		}
		break
	}
	if start == p.pos {
		return nil, fmt.Errorf("jsonx: неожиданный символ %q на позиции %d", p.data[p.pos], p.pos)
	}
	raw := string(p.data[start:p.pos])
	if _, err := strconv.ParseFloat(raw, 64); err != nil {
		return nil, fmt.Errorf("jsonx: некорректное число %q", raw)
	}
	return &Value{Kind: Number, Number: raw}, nil
}

func (p *parser) string() (string, error) {
	if p.data[p.pos] != '"' {
		return "", fmt.Errorf("jsonx: ожидалась строка на позиции %d", p.pos)
	}
	p.pos++
	var buf bytes.Buffer
	for p.pos < len(p.data) {
		c := p.data[p.pos]
		switch c {
		case '"':
			p.pos++
			return buf.String(), nil
		case '\\':
			p.pos++
			if p.pos >= len(p.data) {
				return "", errUnexpectedEnd
			}
			switch esc := p.data[p.pos]; esc {
			case '"', '\\', '/':
				buf.WriteByte(esc)
			case 'b':
				buf.WriteByte('\b')
			case 'f':
				buf.WriteByte('\f')
			case 'n':
				buf.WriteByte('\n')
			case 'r':
				buf.WriteByte('\r')
			case 't':
				buf.WriteByte('\t')
			case 'u':
				if p.pos+4 >= len(p.data) {
					return "", errUnexpectedEnd
				}
				code, err := strconv.ParseUint(string(p.data[p.pos+1:p.pos+5]), 16, 32)
				if err != nil {
					return "", fmt.Errorf("jsonx: некорректная escape-последовательность")
				}
				p.pos += 4
				r := rune(code)
				if r >= 0xD800 && r <= 0xDBFF {
					// суррогатная пара
					if p.pos+6 < len(p.data) && p.data[p.pos+1] == '\\' && p.data[p.pos+2] == 'u' {
						low, err := strconv.ParseUint(string(p.data[p.pos+3:p.pos+7]), 16, 32)
						if err == nil && low >= 0xDC00 && low <= 0xDFFF {
							r = ((r - 0xD800) << 10 | (rune(low) - 0xDC00)) + 0x10000
							p.pos += 6
						}
					}
				}
				buf.WriteRune(r)
			default:
				return "", fmt.Errorf("jsonx: неизвестная escape-последовательность \\%c", esc)
			}
			p.pos++
		default:
			buf.WriteByte(c)
			p.pos++
		}
	}
	return "", errUnexpectedEnd
}

func (p *parser) array() (*Value, error) {
	p.pos++ // [
	v := &Value{Kind: Array}
	p.skipSpace()
	if p.pos < len(p.data) && p.data[p.pos] == ']' {
		p.pos++
		return v, nil
	}
	for {
		p.skipSpace()
		item, err := p.value()
		if err != nil {
			return nil, err
		}
		v.Items = append(v.Items, item)
		p.skipSpace()
		if p.pos >= len(p.data) {
			return nil, errUnexpectedEnd
		}
		switch p.data[p.pos] {
		case ',':
			p.pos++
		case ']':
			p.pos++
			return v, nil
		default:
			return nil, fmt.Errorf("jsonx: ожидался ',' или ']' на позиции %d", p.pos)
		}
	}
}

func (p *parser) object() (*Value, error) {
	p.pos++ // {
	v := &Value{Kind: Object}
	p.skipSpace()
	if p.pos < len(p.data) && p.data[p.pos] == '}' {
		p.pos++
		return v, nil
	}
	for {
		p.skipSpace()
		key, err := p.string()
		if err != nil {
			return nil, err
		}
		p.skipSpace()
		if p.pos >= len(p.data) || p.data[p.pos] != ':' {
			return nil, fmt.Errorf("jsonx: ожидался ':' на позиции %d", p.pos)
		}
		p.pos++
		p.skipSpace()
		val, err := p.value()
		if err != nil {
			return nil, err
		}
		v.Keys = append(v.Keys, key)
		v.Values = append(v.Values, val)
		p.skipSpace()
		if p.pos >= len(p.data) {
			return nil, errUnexpectedEnd
		}
		switch p.data[p.pos] {
		case ',':
			p.pos++
		case '}':
			p.pos++
			return v, nil
		default:
			return nil, fmt.Errorf("jsonx: ожидался ',' или '}' на позиции %d", p.pos)
		}
	}
}

// Marshal сериализует узел компактно, как это делает FastAPI.
func Marshal(v *Value) []byte {
	var buf bytes.Buffer
	writeValue(&buf, v)
	return buf.Bytes()
}

func writeValue(buf *bytes.Buffer, v *Value) {
	if v == nil {
		buf.WriteString("null")
		return
	}
	switch v.Kind {
	case Null:
		buf.WriteString("null")
	case Bool:
		if v.Bool {
			buf.WriteString("true")
		} else {
			buf.WriteString("false")
		}
	case Number:
		buf.WriteString(v.Number)
	case String:
		writeString(buf, v.Str)
	case Array:
		buf.WriteByte('[')
		for i, item := range v.Items {
			if i > 0 {
				buf.WriteByte(',')
			}
			writeValue(buf, item)
		}
		buf.WriteByte(']')
	case Object:
		buf.WriteByte('{')
		for i, k := range v.Keys {
			if i > 0 {
				buf.WriteByte(',')
			}
			writeString(buf, k)
			buf.WriteByte(':')
			writeValue(buf, v.Values[i])
		}
		buf.WriteByte('}')
	}
}

const hexDigits = "0123456789abcdef"

func writeString(buf *bytes.Buffer, s string) {
	buf.WriteByte('"')
	for _, r := range s {
		switch r {
		case '"':
			buf.WriteString(`\"`)
		case '\\':
			buf.WriteString(`\\`)
		case '\n':
			buf.WriteString(`\n`)
		case '\r':
			buf.WriteString(`\r`)
		case '\t':
			buf.WriteString(`\t`)
		case '\b':
			buf.WriteString(`\b`)
		case '\f':
			buf.WriteString(`\f`)
		default:
			if r < 0x20 {
				buf.WriteString(`\u00`)
				buf.WriteByte(hexDigits[byte(r)>>4])
				buf.WriteByte(hexDigits[byte(r)&0xF])
				continue
			}
			buf.WriteRune(r)
		}
	}
	buf.WriteByte('"')
}

// NewInt создаёт числовой узел из целого.
func NewInt(n int64) *Value { return &Value{Kind: Number, Number: strconv.FormatInt(n, 10)} }

// NewFloat создаёт числовой узел из float64 в записи, совпадающей с Python.
func NewFloat(f float64) *Value { return &Value{Kind: Number, Number: FormatFloat(f)} }

// NewBool создаёт булев узел.
func NewBool(b bool) *Value { return &Value{Kind: Bool, Bool: b} }

// NewNull создаёт узел null.
func NewNull() *Value { return &Value{Kind: Null} }

// FormatFloat повторяет формат repr() в Python для чисел с плавающей точкой.
//
// Python печатает 1507500.0, а не 1.5075e+06, и переключается на экспоненту
// только при decpt <= -4 или decpt > 16. Стандартный strconv так не умеет,
// поэтому формат собирается вручную: иначе паритет с mock API по байтам
// ломается на первом же числе сметы.
func FormatFloat(f float64) string {
	switch {
	case math.IsNaN(f) || math.IsInf(f, 0):
		return "null"
	case f == 0:
		if math.Signbit(f) {
			return "-0.0"
		}
		return "0.0"
	}

	// "1.5075e+06" → digits="15075", exp=6
	raw := strconv.FormatFloat(f, 'e', -1, 64)
	mantissa, expPart, _ := strings.Cut(raw, "e")
	exp, err := strconv.Atoi(expPart)
	if err != nil {
		return raw
	}

	negative := strings.HasPrefix(mantissa, "-")
	mantissa = strings.TrimPrefix(mantissa, "-")
	digits := strings.Replace(mantissa, ".", "", 1)

	// decpt — позиция десятичной точки относительно digits.
	decpt := exp + 1

	var out string
	if decpt <= -4 || decpt > 16 {
		out = digits[:1]
		if len(digits) > 1 {
			out += "." + digits[1:]
		}
		sign := "+"
		e := decpt - 1
		if e < 0 {
			sign = "-"
			e = -e
		}
		expDigits := strconv.Itoa(e)
		if len(expDigits) < 2 {
			expDigits = "0" + expDigits
		}
		out += "e" + sign + expDigits
	} else if decpt <= 0 {
		out = "0." + strings.Repeat("0", -decpt) + digits
	} else if decpt >= len(digits) {
		out = digits + strings.Repeat("0", decpt-len(digits)) + ".0"
	} else {
		out = digits[:decpt] + "." + digits[decpt:]
	}

	if negative {
		return "-" + out
	}
	return out
}
