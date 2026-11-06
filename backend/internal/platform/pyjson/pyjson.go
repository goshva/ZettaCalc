// Package pyjson — JSON в семантике Python, а не Go.
//
// Паритет с mock API (FastAPI + pydantic) требует не «правильного», а
// совпадающего поведения: тексты и позиции ошибок json.loads, повторная
// сериализация уже разобранного значения (int остаётся int, float печатается
// как repr), арифметика Python int(). Всё это отличается от encoding/json,
// поэтому пакет реализует собственные разбор и запись.
package pyjson

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Kind — тип значения после разбора.
type Kind uint8

// Виды значений.
const (
	Null Kind = iota
	Bool
	Int
	Float
	String
	Array
	Object
)

// Value — разобранное значение.
type Value struct {
	Kind Kind

	BoolValue bool
	IntValue  *big.Int // для Kind == Int
	FloatVal  float64  // для Kind == Float
	Str       string
	Items     []*Value
	Keys      []string
	Values    []*Value
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

// Field возвращает значение поля или nil: отсутствие поля означает «по умолчанию».
func (v *Value) Field(key string) *Value {
	item, _ := v.Get(key)
	return item
}

// IsNull сообщает, что узел — null или отсутствует.
func (v *Value) IsNull() bool { return v == nil || v.Kind == Null }

// IsString сообщает, что узел — строка.
func (v *Value) IsString() bool { return v != nil && v.Kind == String }

// SyntaxError — ошибка разбора с позицией в символах, как у json.loads.
type SyntaxError struct {
	// Pos — индекс символа (не байта), как в Python JSONDecodeError.pos.
	Pos int
	// Msg — текст ошибки из json.loads (JSONDecodeError.msg).
	Msg string
}

func (e *SyntaxError) Error() string {
	return fmt.Sprintf("pyjson: %s (позиция %d)", e.Msg, e.Pos)
}

// NewString создаёт строковое значение.
func NewString(s string) *Value { return &Value{Kind: String, Str: s} }

// NewFloat создаёт дробное значение.
func NewFloat(f float64) *Value { return &Value{Kind: Float, FloatVal: f} }

// NewObject создаёт пустой объект.
func NewObject() *Value { return &Value{Kind: Object} }

// NewArray создаёт массив.
func NewArray(items ...*Value) *Value { return &Value{Kind: Array, Items: items} }

// Set добавляет поле объекта.
func (v *Value) Set(key string, value *Value) {
	v.Keys = append(v.Keys, key)
	v.Values = append(v.Values, value)
}

// Repr повторяет repr() из Python.
//
// Нужен там, где mock API строит значение из самого тела запроса: например,
// идентификатор ЦПР считается как md5(str(payload)), то есть от repr словаря.
func Repr(v *Value) string {
	var buf bytes.Buffer
	writeRepr(&buf, v)
	return buf.String()
}

func writeRepr(buf *bytes.Buffer, v *Value) {
	if v == nil {
		buf.WriteString("None")
		return
	}
	switch v.Kind {
	case Null:
		buf.WriteString("None")
	case Bool:
		if v.BoolValue {
			buf.WriteString("True")
		} else {
			buf.WriteString("False")
		}
	case Int:
		buf.WriteString(v.IntValue.String())
	case Float:
		buf.WriteString(reprFloat(v.FloatVal))
	case String:
		writeReprString(buf, v.Str)
	case Array:
		buf.WriteByte('[')
		for i, item := range v.Items {
			if i > 0 {
				buf.WriteString(", ")
			}
			writeRepr(buf, item)
		}
		buf.WriteByte(']')
	case Object:
		buf.WriteByte('{')
		for i, k := range v.Keys {
			if i > 0 {
				buf.WriteString(", ")
			}
			writeReprString(buf, k)
			buf.WriteString(": ")
			writeRepr(buf, v.Values[i])
		}
		buf.WriteByte('}')
	}
}

func reprFloat(f float64) string {
	switch {
	case math.IsInf(f, 1):
		return "inf"
	case math.IsInf(f, -1):
		return "-inf"
	case math.IsNaN(f):
		return "nan"
	}
	return FormatFloat(f)
}

// writeReprString повторяет выбор кавычек в Python: одинарные, если строка не
// содержит одинарную кавычку.
func writeReprString(buf *bytes.Buffer, s string) {
	quote := byte('\'')
	if strings.ContainsRune(s, '\'') && !strings.ContainsRune(s, '"') {
		quote = '"'
	}
	buf.WriteByte(quote)
	for _, r := range s {
		switch r {
		case '\\':
			buf.WriteString(`\\`)
		case '\n':
			buf.WriteString(`\n`)
		case '\r':
			buf.WriteString(`\r`)
		case '\t':
			buf.WriteString(`\t`)
		default:
			switch {
			case r == rune(quote):
				buf.WriteByte('\\')
				buf.WriteRune(r)
			case r < 0x20 || r == 0x7f:
				fmt.Fprintf(buf, `\x%02x`, r)
			default:
				buf.WriteRune(r)
			}
		}
	}
	buf.WriteByte(quote)
}

// NewInt создаёт целое значение.
func NewInt(n int64) *Value { return &Value{Kind: Int, IntValue: big.NewInt(n)} }

// NewBool создаёт булево значение.
func NewBool(b bool) *Value { return &Value{Kind: Bool, BoolValue: b} }

// NewNull создаёт null.
func NewNull() *Value { return &Value{Kind: Null} }

// Decode разбирает JSON и повторяет ошибки json.loads.
func Decode(data []byte) (*Value, error) {
	runes := []rune(string(data))
	p := &parser{src: runes}
	p.skipSpace()
	if p.pos >= len(p.src) {
		return nil, &SyntaxError{Pos: p.pos, Msg: "Expecting value"}
	}
	v, err := p.value()
	if err != nil {
		return nil, err
	}
	p.skipSpace()
	if p.pos != len(p.src) {
		return nil, &SyntaxError{Pos: p.pos, Msg: "Extra data"}
	}
	return v, nil
}

type parser struct {
	src []rune
	pos int
}

func (p *parser) skipSpace() {
	for p.pos < len(p.src) {
		switch p.src[p.pos] {
		case ' ', '\t', '\n', '\r':
			p.pos++
		default:
			return
		}
	}
}

func (p *parser) errAt(pos int, msg string) error { return &SyntaxError{Pos: pos, Msg: msg} }

func (p *parser) value() (*Value, error) {
	if p.pos >= len(p.src) {
		return nil, p.errAt(p.pos, "Expecting value")
	}
	switch c := p.src[p.pos]; {
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
		return p.literal("true", true)
	case c == 'f':
		return p.literal("false", false)
	case c == 'n':
		if _, err := p.literalRaw("null"); err != nil {
			return nil, err
		}
		return &Value{Kind: Null}, nil
	case c == 'N':
		// json.loads принимает литерал NaN, хотя JSON его не разрешает.
		return p.nonFinite("NaN", math.NaN())
	case c == 'I':
		return p.nonFinite("Infinity", math.Inf(1))
	case c == '-':
		if p.hasPrefix("-Infinity") {
			return p.nonFinite("-Infinity", math.Inf(-1))
		}
		return p.number()
	case c >= '0' && c <= '9':
		return p.number()
	default:
		return nil, p.errAt(p.pos, "Expecting value")
	}
}

// hasPrefix сообщает, что в текущей позиции стоит указанный литерал.
func (p *parser) hasPrefix(lit string) bool {
	runes := []rune(lit)
	if len(p.src)-p.pos < len(runes) {
		return false
	}
	for i, r := range runes {
		if p.src[p.pos+i] != r {
			return false
		}
	}
	return true
}

// nonFinite разбирает литералы Infinity и NaN.
//
// json.loads принимает их без allow_nan=False, но сериализация ответа с таким
// значением падает — mock API отвечает 500 (см. httpx.WriteInternalError).
func (p *parser) nonFinite(lit string, value float64) (*Value, error) {
	if _, err := p.literalRaw(lit); err != nil {
		return nil, err
	}
	return &Value{Kind: Float, FloatVal: value}, nil
}

func (p *parser) literalRaw(lit string) (bool, error) {
	start := p.pos
	runes := []rune(lit)
	if len(p.src)-p.pos < len(runes) {
		return false, p.errAt(start, "Expecting value")
	}
	for i, r := range runes {
		if p.src[p.pos+i] != r {
			return false, p.errAt(start, "Expecting value")
		}
	}
	p.pos += len(runes)
	return true, nil
}

func (p *parser) literal(lit string, value bool) (*Value, error) {
	if _, err := p.literalRaw(lit); err != nil {
		return nil, err
	}
	return &Value{Kind: Bool, BoolValue: value}, nil
}

// number повторяет разбор чисел в CPython: некорректный «хвост» числа
// (например, «01» или «1.») не считается ошибкой числа — значение обрывается,
// а ошибку выдаёт вызывающий контекст.
func (p *parser) number() (*Value, error) {
	start := p.pos
	isFloat := false
	if p.pos < len(p.src) && p.src[p.pos] == '-' {
		p.pos++
	}
	p.digits()
	if p.pos < len(p.src) && p.src[p.pos] == '.' {
		next := p.pos + 1
		if next < len(p.src) && isDigit(p.src[next]) {
			p.pos = next
			p.digits()
			isFloat = true
		}
	}
	if p.pos < len(p.src) && (p.src[p.pos] == 'e' || p.src[p.pos] == 'E') {
		next := p.pos + 1
		if next < len(p.src) && (p.src[next] == '+' || p.src[next] == '-') {
			next++
		}
		if next < len(p.src) && isDigit(p.src[next]) {
			p.pos = next
			p.digits()
			isFloat = true
		}
	}
	raw := string(p.src[start:p.pos])
	if raw == "" || raw == "-" {
		return nil, p.errAt(start, "Expecting value")
	}
	if isFloat {
		f, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			if !errors.Is(err, strconv.ErrRange) {
				return nil, p.errAt(start, "Expecting value")
			}
			// Переполнение даёт ±Inf, а не ошибку: json.loads("1e400") == inf.
			// Такое значение допустимо при разборе, но ломает сериализацию
			// ответа (см. HasNonFinite).
			return &Value{Kind: Float, FloatVal: f}, nil
		}
		return &Value{Kind: Float, FloatVal: f}, nil
	}
	n, ok := new(big.Int).SetString(raw, 10)
	if !ok {
		return nil, p.errAt(start, "Expecting value")
	}
	return &Value{Kind: Int, IntValue: n}, nil
}

func (p *parser) digits() {
	for p.pos < len(p.src) && isDigit(p.src[p.pos]) {
		p.pos++
	}
}

func isDigit(r rune) bool { return r >= '0' && r <= '9' }

func (p *parser) object() (*Value, error) {
	p.pos++ // {
	v := &Value{Kind: Object}
	p.skipSpace()
	if p.pos < len(p.src) && p.src[p.pos] == '}' {
		p.pos++
		return v, nil
	}
	for {
		p.skipSpace()
		if p.pos >= len(p.src) || p.src[p.pos] != '"' {
			return nil, p.errAt(p.pos, "Expecting property name enclosed in double quotes")
		}
		key, err := p.string()
		if err != nil {
			return nil, err
		}
		p.skipSpace()
		if p.pos >= len(p.src) || p.src[p.pos] != ':' {
			return nil, p.errAt(p.pos, "Expecting ':' delimiter")
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
		if p.pos >= len(p.src) {
			return nil, p.errAt(p.pos, "Expecting ',' delimiter")
		}
		switch p.src[p.pos] {
		case ',':
			p.pos++
		case '}':
			p.pos++
			return v, nil
		default:
			return nil, p.errAt(p.pos, "Expecting ',' delimiter")
		}
	}
}

func (p *parser) array() (*Value, error) {
	p.pos++ // [
	v := &Value{Kind: Array}
	p.skipSpace()
	if p.pos < len(p.src) && p.src[p.pos] == ']' {
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
		if p.pos >= len(p.src) {
			return nil, p.errAt(p.pos, "Expecting ',' delimiter")
		}
		switch p.src[p.pos] {
		case ',':
			p.pos++
		case ']':
			p.pos++
			return v, nil
		default:
			return nil, p.errAt(p.pos, "Expecting ',' delimiter")
		}
	}
}

func (p *parser) string() (string, error) {
	start := p.pos
	p.pos++ // открывающая кавычка
	var buf bytes.Buffer
	for p.pos < len(p.src) {
		c := p.src[p.pos]
		switch {
		case c == '"':
			p.pos++
			return buf.String(), nil
		case c == '\\':
			p.pos++
			if p.pos >= len(p.src) {
				return "", p.errAt(start, "Unterminated string starting at")
			}
			switch esc := p.src[p.pos]; esc {
			case '"', '\\', '/':
				buf.WriteRune(esc)
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
				r, err := p.unicodeEscape()
				if err != nil {
					return "", err
				}
				buf.WriteRune(r)
			default:
				return "", p.errAt(p.pos-1, fmt.Sprintf("Invalid \\escape: %c", esc))
			}
			p.pos++
		case c < 0x20:
			return "", p.errAt(p.pos, "Invalid control character at")
		default:
			buf.WriteRune(c)
			p.pos++
		}
	}
	return "", p.errAt(start, "Unterminated string starting at")
}

func (p *parser) unicodeEscape() (rune, error) {
	code, err := p.hex4()
	if err != nil {
		return 0, err
	}
	r := rune(code)
	if r >= 0xD800 && r <= 0xDBFF {
		if p.pos+6 < len(p.src) && p.src[p.pos+1] == '\\' && p.src[p.pos+2] == 'u' {
			p.pos += 2
			low, err := p.hex4()
			if err != nil {
				return 0, err
			}
			if low >= 0xDC00 && low <= 0xDFFF {
				return ((r - 0xD800) << 10 | (rune(low) - 0xDC00)) + 0x10000, nil
			}
			p.pos -= 2
		}
	}
	return r, nil
}

func (p *parser) hex4() (uint32, error) {
	if p.pos+4 >= len(p.src) {
		return 0, p.errAt(p.pos, "Invalid \\uXXXX escape")
	}
	hex := string(p.src[p.pos+1 : p.pos+5])
	code, err := strconv.ParseUint(hex, 16, 32)
	if err != nil {
		return 0, p.errAt(p.pos, "Invalid \\uXXXX escape")
	}
	p.pos += 4
	return uint32(code), nil
}

// Encode записывает значение так, как это делает json.dumps после разбора.
//
// Числа нормализуются: целое печатается целым, дробное — repr() из Python.
// Так ведёт себя FastAPI, возвращающий уже разобранное тело запроса.
func Encode(v *Value) []byte {
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
		if v.BoolValue {
			buf.WriteString("true")
		} else {
			buf.WriteString("false")
		}
	case Int:
		buf.WriteString(v.IntValue.String())
	case Float:
		buf.WriteString(FormatFloat(v.FloatVal))
	case String:
		WriteString(buf, v.Str)
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
			WriteString(buf, k)
			buf.WriteByte(':')
			writeValue(buf, v.Values[i])
		}
		buf.WriteByte('}')
	}
}

const hexDigits = "0123456789abcdef"

// FormatNumber печатает число так, как это сделал бы Python, если бы значение
// пришло целым: 385000 вместо 385000.0. Нужно для полей, которые в mock API
// заданы целыми (например, цена позиции корзины), но в БД хранятся как numeric.
func FormatNumber(f float64) string {
	if !math.IsInf(f, 0) && !math.IsNaN(f) && f == math.Trunc(f) && math.Abs(f) < 1e15 {
		return strconv.FormatInt(int64(f), 10)
	}
	return FormatFloat(f)
}

// WriteString пишет строку так же, как json.dumps(ensure_ascii=False).
func WriteString(buf *bytes.Buffer, s string) {
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

// FormatFloat повторяет repr() чисел с плавающей точкой в Python.
// Python печатает 1507500.0, а не 1.5075e+06, и переключается на экспоненту
// только при decpt <= -4 или decpt > 16; strconv.FormatFloat так не умеет.
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

	raw := strconv.FormatFloat(f, 'e', -1, 64)
	mantissa, expPart, _ := strings.Cut(raw, "e")
	exp, err := strconv.Atoi(expPart)
	if err != nil {
		return raw
	}

	negative := strings.HasPrefix(mantissa, "-")
	mantissa = strings.TrimPrefix(mantissa, "-")
	digits := strings.Replace(mantissa, ".", "", 1)
	decpt := exp + 1

	var out string
	switch {
	case decpt <= -4 || decpt > 16:
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
	case decpt <= 0:
		out = "0." + strings.Repeat("0", -decpt) + digits
	case decpt >= len(digits):
		out = digits + strings.Repeat("0", decpt-len(digits)) + ".0"
	default:
		out = digits[:decpt] + "." + digits[decpt:]
	}

	if negative {
		return "-" + out
	}
	return out
}

// AsInt повторяет int() из Python.
//
// Это не «безопасное приведение»: int("1_0") == 10, int(" 5 ") == 5,
// int(2.7) == 2, int(True) == 1, а int("abc") — ошибка TypeError, которую
// mock API отдаёт как 500 Internal Server Error.
func AsInt(v *Value) (*big.Int, error) {
	if v == nil {
		return nil, fmt.Errorf("pyjson: int(None) недопустим")
	}
	switch v.Kind {
	case Int:
		return new(big.Int).Set(v.IntValue), nil
	case Float:
		// int(inf) и int(nan) в Python — OverflowError/ValueError, то есть 500.
		if math.IsInf(v.FloatVal, 0) || math.IsNaN(v.FloatVal) {
			return nil, fmt.Errorf("pyjson: int() от неконечного числа недопустим")
		}
		f := math.Trunc(v.FloatVal)
		result, _ := new(big.Float).SetFloat64(f).Int(nil)
		return result, nil
	case Bool:
		if v.BoolValue {
			return big.NewInt(1), nil
		}
		return big.NewInt(0), nil
	case String:
		return parseIntString(v.Str)
	default:
		return nil, fmt.Errorf("pyjson: int() от %v недопустим", v.Kind)
	}
}

func parseIntString(s string) (*big.Int, error) {
	trimmed := strings.TrimFunc(s, unicode.IsSpace)
	if trimmed == "" {
		return nil, fmt.Errorf("pyjson: некорректный литерал для int()")
	}
	negative := false
	switch trimmed[0] {
	case '+':
		trimmed = trimmed[1:]
	case '-':
		negative = true
		trimmed = trimmed[1:]
	}
	if trimmed == "" {
		return nil, fmt.Errorf("pyjson: некорректный литерал для int()")
	}
	cleaned := strings.ReplaceAll(trimmed, "_", "")
	if cleaned == "" {
		return nil, fmt.Errorf("pyjson: некорректный литерал для int()")
	}
	for _, r := range cleaned {
		if !isDigit(r) {
			return nil, fmt.Errorf("pyjson: некорректный литерал для int()")
		}
	}
	n, ok := new(big.Int).SetString(cleaned, 10)
	if !ok {
		return nil, fmt.Errorf("pyjson: некорректный литерал для int()")
	}
	if negative {
		n.Neg(n)
	}
	return n, nil
}

// RuneLen возвращает длину строки в символах (как len(str) в Python).
func RuneLen(s string) int { return utf8.RuneCountInString(s) }

// HasNonFinite сообщает, что в значении есть NaN или ±Inf.
//
// json.loads такие литералы принимает, но json.dumps без allow_nan падает на
// них ValueError, поэтому mock API отдаёт 500, если non-finite попало в тело
// ответа (эхо тела запроса или значение в ошибке валидации).
func HasNonFinite(v *Value) bool {
	if v == nil {
		return false
	}
	switch v.Kind {
	case Float:
		return math.IsNaN(v.FloatVal) || math.IsInf(v.FloatVal, 0)
	case Array:
		for _, item := range v.Items {
			if HasNonFinite(item) {
				return true
			}
		}
	case Object:
		for _, item := range v.Values {
			if HasNonFinite(item) {
				return true
			}
		}
	}
	return false
}
