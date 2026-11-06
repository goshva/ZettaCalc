// Package pybody — чтение тела запроса так, как это делает FastAPI.
//
// Порядок и результат разбора важны для паритета:
//  1. тело читается целиком;
//  2. если content-type «json-подобный», байты декодируются в UTF-8 (с учётом
//     BOM и UTF-16/32); ошибка декодирования → 400 «There was an error parsing
//     the body»;
//  3. тело разбирается json.loads; ошибка разбора → 422 с типом json_invalid и
//     позицией в символах;
//  4. если content-type не json-подобный, тело передаётся как строка, и
//     pydantic отвечает model_attributes_type / dict_type;
//  5. пустое тело → ошибка missing по loc ["body"].
package pybody

import (
	"encoding/binary"
	"io"
	"mime"
	"strings"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/zetta/zettacalc/backend/internal/platform/pyjson"
	"github.com/zetta/zettacalc/backend/internal/platform/pyvalidate"
)

// Limit — максимальный размер читаемого тела.
const Limit = 4 << 20

// Result — результат разбора тела.
type Result struct {
	// Value — разобранное значение (nil, если тела нет).
	Value *pyjson.Value
	// Raw — исходные байты тела.
	Raw []byte
	// Empty — тело пустое.
	Empty bool
	// Err — ошибка чтения (например, некорректная кодировка).
	Err *pyvalidate.Error
	// DecodeError — тело не удалось декодировать как UTF-8.
	DecodeError bool
}

// Read читает и разбирает тело запроса.
func Read(r io.Reader, contentType string) Result {
	raw, err := io.ReadAll(io.LimitReader(r, Limit))
	if err != nil {
		return Result{DecodeError: true}
	}
	if len(raw) == 0 {
		return Result{Raw: raw, Empty: true}
	}

	if !IsJSON(contentType) {
		// Тело не JSON: pydantic получит строку и пожалуется на её тип.
		return Result{Value: pyjson.NewString(string(raw)), Raw: raw}
	}

	text, ok := decodeText(raw)
	if !ok {
		return Result{Raw: raw, DecodeError: true}
	}
	value, err := pyjson.Decode([]byte(text))
	if err != nil {
		syntax, _ := err.(*pyjson.SyntaxError)
		msg := "Expecting value"
		pos := 0
		if syntax != nil {
			msg = syntax.Msg
			pos = syntax.Pos
		}
		return Result{Raw: raw, Err: &pyvalidate.Error{
			Type:  "json_invalid",
			Loc:   pyvalidate.Body(pos),
			Msg:   "JSON decode error",
			Input: pyjson.NewObject(),
			Ctx:   pyvalidate.NewCtx().SetString("error", msg),
		}}
	}
	return Result{Value: value, Raw: raw}
}

// IsJSON сообщает, что content-type разбирается как JSON.
//
// FastAPI считает JSON-совместимыми application/json и суффикс +json,
// а также отсутствие content-type.
func IsJSON(contentType string) bool {
	if strings.TrimSpace(contentType) == "" {
		return false
	}
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		mediaType = contentType
	}
	mediaType = strings.ToLower(strings.TrimSpace(mediaType))
	if mediaType == "application/json" {
		return true
	}
	return strings.HasSuffix(mediaType, "+json")
}

// decodeText декодирует байты в строку: UTF-8 (в том числе с BOM), UTF-16 и
// UTF-32 — как это делает json.loads для bytes.
func decodeText(raw []byte) (string, bool) {
	switch {
	case len(raw) >= 4 && raw[0] == 0x00 && raw[1] == 0x00 && raw[2] == 0xFE && raw[3] == 0xFF:
		return decodeUTF32(raw[4:], binary.BigEndian)
	case len(raw) >= 4 && raw[0] == 0xFF && raw[1] == 0xFE && raw[2] == 0x00 && raw[3] == 0x00:
		return decodeUTF32(raw[4:], binary.LittleEndian)
	case len(raw) >= 2 && raw[0] == 0xFF && raw[1] == 0xFE:
		return decodeUTF16(raw[2:], binary.LittleEndian)
	case len(raw) >= 2 && raw[0] == 0xFE && raw[1] == 0xFF:
		return decodeUTF16(raw[2:], binary.BigEndian)
	}

	text := raw
	// UTF-8 BOM снимается до разбора, но позиции ошибок считаются по строке
	// без BOM — так же, как это делает json.loads(bytes).
	if len(text) >= 3 && text[0] == 0xEF && text[1] == 0xBB && text[2] == 0xBF {
		text = text[3:]
	}
	if !utf8.Valid(text) {
		return "", false
	}
	return string(text), true
}

func decodeUTF16(raw []byte, order binary.ByteOrder) (string, bool) {
	if len(raw)%2 != 0 {
		return "", false
	}
	units := make([]uint16, 0, len(raw)/2)
	for i := 0; i < len(raw); i += 2 {
		units = append(units, order.Uint16(raw[i:i+2]))
	}
	return string(utf16.Decode(units)), true
}

func decodeUTF32(raw []byte, order binary.ByteOrder) (string, bool) {
	if len(raw)%4 != 0 {
		return "", false
	}
	var sb strings.Builder
	for i := 0; i < len(raw); i += 4 {
		sb.WriteRune(rune(order.Uint32(raw[i : i+4])))
	}
	return sb.String(), true
}

// MissingBody — ошибка pydantic для отсутствующего тела.
func MissingBody() *pyvalidate.Error {
	return &pyvalidate.Error{
		Type:  "missing",
		Loc:   pyvalidate.Body(),
		Msg:   "Field required",
		Input: pyjson.NewNull(),
	}
}

// NotADict — ошибка pydantic, когда тело не является объектом.
//
// Для dict-моделей FastAPI использует тип dict_type и текст без упоминания
// полей; для pydantic-моделей — model_attributes_type.
func NotADict(value *pyjson.Value, forModel bool) *pyvalidate.Error {
	typ := "dict_type"
	msg := "Input should be a valid dictionary"
	if forModel {
		typ = "model_attributes_type"
		msg = "Input should be a valid dictionary or object to extract fields from"
	}
	return &pyvalidate.Error{
		Type:  typ,
		Loc:   pyvalidate.Body(),
		Msg:   msg,
		Input: value,
	}
}

// IsObject сообщает, что значение — JSON-объект.
func IsObject(v *pyjson.Value) bool { return v != nil && v.Kind == pyjson.Object }
