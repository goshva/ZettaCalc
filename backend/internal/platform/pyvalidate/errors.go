// Package pyvalidate — валидация входных данных в семантике pydantic v2.
//
// Зачем: mock API построен на FastAPI, поэтому его ответы на некорректный
// запрос — это не «ошибка 400», а массив ошибок вида
//
//	{"detail":[{"type":"string_too_short","loc":["body","name"],"msg":"…","input":"","ctx":{"min_length":1}}]}
//
// Фронтенды прототипов показывают эти тексты пользователю, значит их нужно
// воспроизводить дословно. Пакет отвечает только за форму ошибок и приведения
// типов (lax-режим pydantic); доменные правила остаются в контекстах.
package pyvalidate

import (
	"bytes"
	"strconv"

	"github.com/zetta/zettacalc/backend/internal/platform/pyjson"
)

// Error — одна ошибка в формате pydantic.
type Error struct {
	Type  string
	Loc   []any
	Msg   string
	Input *pyjson.Value
	Ctx   *Ctx
}

// Ctx — дополнительный контекст ошибки (порядок ключей сохраняется).
type Ctx struct {
	keys   []string
	values []*pyjson.Value
}

// NewCtx создаёт пустой контекст.
func NewCtx() *Ctx { return &Ctx{} }

// Set добавляет поле контекста.
func (c *Ctx) Set(key string, value *pyjson.Value) *Ctx {
	c.keys = append(c.keys, key)
	c.values = append(c.values, value)
	return c
}

// SetString добавляет строковое поле контекста.
func (c *Ctx) SetString(key, value string) *Ctx { return c.Set(key, pyjson.NewString(value)) }

// SetInt добавляет целое поле контекста.
func (c *Ctx) SetInt(key string, value int) *Ctx {
	return c.Set(key, pyjson.NewInt(int64(value)))
}

// Body возвращает loc вида ["body", …].
func Body(parts ...any) []any { return append([]any{"body"}, parts...) }

// Query возвращает loc вида ["query", …].
func Query(parts ...any) []any { return append([]any{"query"}, parts...) }

// Errors — набор ошибок запроса.
type Errors struct {
	items []Error
}

// Add добавляет ошибку.
func (e *Errors) Add(err Error) { e.items = append(e.items, err) }

// Len возвращает число ошибок.
func (e *Errors) Len() int { return len(e.items) }

// Items возвращает ошибки.
func (e *Errors) Items() []Error { return e.items }

// Body — тело ответа 422: {"detail":[…]}, как у FastAPI.
func (e *Errors) Body() []byte { return encodeBody(e.items) }

// ResponseBody — тело ответа 422 с учётом приоритета ошибок источников.
//
// FastAPI разбирает тело запроса до валидации query-параметров: если json.loads
// упал, до параметров дело не доходит и их ошибки в ответ не попадают. Поэтому
// json_invalid вытесняет все прочие ошибки.
func (e *Errors) ResponseBody() []byte {
	for _, item := range e.items {
		if item.Type == "json_invalid" {
			return encodeBody([]Error{item})
		}
	}
	return e.Body()
}

func encodeBody(items []Error) []byte {
	var buf bytes.Buffer
	buf.WriteString(`{"detail":[`)
	for i, item := range items {
		if i > 0 {
			buf.WriteByte(',')
		}
		writeError(&buf, item)
	}
	buf.WriteString(`]}`)
	return buf.Bytes()
}

// HasNonFinite сообщает, что в ошибках есть значение NaN или ±Inf.
//
// Такое значение попадает в ответ через поле input, а json.dumps без allow_nan
// на нём падает, поэтому mock API отвечает 500 вместо 422.
func (e *Errors) HasNonFinite() bool {
	for _, item := range e.items {
		if pyjson.HasNonFinite(item.Input) {
			return true
		}
		if item.Ctx != nil {
			for _, value := range item.Ctx.values {
				if pyjson.HasNonFinite(value) {
					return true
				}
			}
		}
	}
	return false
}

func writeError(buf *bytes.Buffer, item Error) {
	buf.WriteString(`{"type":`)
	pyjson.WriteString(buf, item.Type)
	buf.WriteString(`,"loc":[`)
	for i, part := range item.Loc {
		if i > 0 {
			buf.WriteByte(',')
		}
		switch v := part.(type) {
		case string:
			pyjson.WriteString(buf, v)
		case int:
			buf.WriteString(strconv.Itoa(v))
		}
	}
	buf.WriteString(`],"msg":`)
	pyjson.WriteString(buf, item.Msg)
	buf.WriteString(`,"input":`)
	if item.Input == nil {
		buf.WriteString("null")
	} else {
		buf.Write(pyjson.Encode(item.Input))
	}
	if item.Ctx != nil {
		buf.WriteString(`,"ctx":{`)
		for i, key := range item.Ctx.keys {
			if i > 0 {
				buf.WriteByte(',')
			}
			pyjson.WriteString(buf, key)
			buf.WriteByte(':')
			buf.Write(pyjson.Encode(item.Ctx.values[i]))
		}
		buf.WriteByte('}')
	}
	buf.WriteByte('}')
}

// PluralCharacters возвращает слово «character(s)» как в pydantic.
func PluralCharacters(n int) string {
	if n == 1 {
		return "character"
	}
	return "characters"
}

// PluralItems возвращает слово «item(s)» как в pydantic.
func PluralItems(n int) string {
	if n == 1 {
		return "item"
	}
	return "items"
}

// InputBodyNone — input для ошибки отсутствующего тела.
func InputBodyNone() *pyjson.Value { return pyjson.NewNull() }
