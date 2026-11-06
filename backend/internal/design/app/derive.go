// Package app контекста Design: детерминированный вывод сметы, ЦПР и ставок.
//
// Логика повторяет заглушку mock API один в один, включая порядок ключей JSON:
// паритет проверяется сравнением тел ответов. Любое изменение формул —
// изменение контракта, а не «улучшение».
package app

import (
	"context"
	"crypto/md5"
	"encoding/hex"

	"github.com/zetta/zettacalc/backend/internal/design/domain"
	"github.com/zetta/zettacalc/backend/internal/platform/jsonx"
)

// Derive — сценарии, производные от выбранного варианта дизайна.
type Derive struct {
	variants *domain.Service
}

// NewDerive собирает сценарии вывода.
func NewDerive(variants *domain.Service) *Derive { return &Derive{variants: variants} }

// Variant возвращает вариант по идентификатору.
func (d *Derive) Variant(ctx context.Context, id string) (domain.Variant, error) {
	return d.variants.ByID(ctx, id)
}

// EstimateJSON собирает смету варианта.
func (d *Derive) EstimateJSON(ctx context.Context, id string) ([]byte, error) {
	v, err := d.Variant(ctx, id)
	if err != nil {
		return nil, err
	}
	return EstimatePayload(v), nil
}

// CPRJSON собирает ЦПР варианта.
func (d *Derive) CPRJSON(ctx context.Context, id string) ([]byte, error) {
	v, err := d.Variant(ctx, id)
	if err != nil {
		return nil, err
	}
	return CPRPayload(v), nil
}

// BidsJSON собирает ставки подрядчиков по варианту.
func (d *Derive) BidsJSON(ctx context.Context, id string) ([]byte, error) {
	v, err := d.Variant(ctx, id)
	if err != nil {
		return nil, err
	}
	return BidsPayload(v), nil
}

// EstimatePayload формирует смету варианта.
//
// total = (budget_min + budget_max) // 2 — целочисленное деление, как в mock API.
func EstimatePayload(v domain.Variant) []byte {
	total := (v.BudgetMin + v.BudgetMax) / 2

	lines := jsonx.NewArray(
		estimateLine("l1", "Мебель и материалы", "комплект", float64(total)*0.45),
		estimateLine("l2", "Ремонт и отделка", "комплект", float64(total)*0.30),
		estimateLine("l3", "Сантехника и электрика", "комплект", float64(total)*0.10),
		estimateLine("l4", "Декор и аксессуары", "комплект", float64(total)*0.05),
		estimateLine("l5", "Проектирование", "услуга", float64(total)*0.10),
	)

	obj := jsonx.NewObject()
	obj.Set("id", jsonx.NewString("est_"+v.ID))
	obj.Set("variant_id", jsonx.NewString(v.ID))
	obj.Set("variant_name", jsonx.NewString(v.Name))
	obj.Set("status", jsonx.NewString("preliminary"))
	obj.Set("total_rub", jsonx.NewInt(total))
	obj.Set("range_min", jsonx.NewInt(v.BudgetMin))
	obj.Set("range_max", jsonx.NewInt(v.BudgetMax))
	obj.Set("days", jsonx.NewInt(int64(v.EstimatedDays)))
	obj.Set("lines", lines)
	obj.Set("assumptions", jsonx.NewArray(
		jsonx.NewString("Ровность пола требует проверки"),
		jsonx.NewString("Состояние стен неизвестно"),
		jsonx.NewString("Цены указаны ориентировочно"),
	))
	return jsonx.Marshal(obj)
}

func estimateLine(id, title, unit string, price float64) *jsonx.Value {
	line := jsonx.NewObject()
	line.Set("id", jsonx.NewString(id))
	line.Set("title", jsonx.NewString(title))
	line.Set("qty", jsonx.NewInt(1))
	line.Set("unit", jsonx.NewString(unit))
	line.Set("price", jsonx.NewFloat(price))
	return line
}

// CPRPayload формирует ЦПР варианта. Идентификатор ЦПР — md5 от id варианта.
func CPRPayload(v domain.Variant) []byte {
	sum := md5.Sum([]byte(v.ID))
	cprID := "cpr_" + hex.EncodeToString(sum[:])[:8]

	obj := jsonx.NewObject()
	obj.Set("id", jsonx.NewString(cprID))
	obj.Set("version", jsonx.NewInt(1))
	obj.Set("project_id", jsonx.NewString("proj_"+v.ID))
	obj.Set("estimate", mustParse(EstimatePayload(v)))
	obj.Set("documents", jsonx.NewArray(
		document("d1", "Планировочное решение", "drawing", "ready"),
		document("d2", "Ведомость отделки", "specification", "ready"),
		document("d3", "Электрическая схема", "drawing", "ready"),
		document("d4", "Смета на работы", "estimate", "ready"),
		document("d5", "График производства работ", "schedule", "pending_review"),
	))
	obj.Set("status", jsonx.NewString("ready"))
	return jsonx.Marshal(obj)
}

func document(id, title, typ, status string) *jsonx.Value {
	doc := jsonx.NewObject()
	doc.Set("id", jsonx.NewString(id))
	doc.Set("title", jsonx.NewString(title))
	doc.Set("type", jsonx.NewString(typ))
	doc.Set("status", jsonx.NewString(status))
	return doc
}

// BidsPayload формирует ставки подрядчиков.
func BidsPayload(v domain.Variant) []byte {
	total := (v.BudgetMin + v.BudgetMax) / 2
	rooms := joinRooms(v.Rooms)

	items := jsonx.NewArray(
		bid("b1", "РемонтПро", 4.8, int64(float64(total)*0.95), int64(v.EstimatedDays)+5,
			"Полный ремонт: "+rooms, "repair"),
		bid("b2", "СтройМастер", 4.5, int64(float64(total)*0.88), int64(v.EstimatedDays)+10,
			"Чистовой ремонт: "+rooms, "auction"),
		bid("b3", "ДизайнРемонт", 4.9, int64(float64(total)*1.10), int64(v.EstimatedDays)-5,
			"Дизайн + ремонт: "+rooms, "expert"),
	)
	return jsonx.Marshal(items)
}

func bid(id, name string, rating float64, price, days int64, scope, image string) *jsonx.Value {
	item := jsonx.NewObject()
	item.Set("id", jsonx.NewString(id))
	item.Set("name", jsonx.NewString(name))
	item.Set("rating", jsonx.NewFloat(rating))
	item.Set("price", jsonx.NewInt(price))
	item.Set("days", jsonx.NewInt(days))
	item.Set("scope", jsonx.NewString(scope))
	item.Set("image", jsonx.NewString(image))
	return item
}

// joinRooms повторяет ", ".join(rooms) из mock API.
func joinRooms(rooms []string) string {
	out := ""
	for i, r := range rooms {
		if i > 0 {
			out += ", "
		}
		out += r
	}
	return out
}

func mustParse(b []byte) *jsonx.Value {
	v, err := jsonx.Parse(b)
	if err != nil {
		panic("design: некорректный JSON сметы: " + err.Error())
	}
	return v
}
