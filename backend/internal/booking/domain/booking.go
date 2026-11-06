// Package domain контекста Booking: запись в Интерьерный центр.
//
// Это единственная write-операция mock API. Правила проекта соблюдаются буквально:
// проверка формы только кодом (никакой LLM), запись — только с Idempotency-Key,
// сначала dry-run, затем подтверждение (approval-gate).
package domain

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"
)

// Ошибки уровня домена.
var (
	// ErrConsentRequired — не подтверждено согласие на обработку данных.
	ErrConsentRequired = errors.New("Не подтверждено согласие на обработку персональных данных")
	// ErrNotFound — запись отсутствует.
	ErrNotFound = errors.New("booking: запись не найдена")
)

// ConsentRequiredMessage — текст ошибки, совпадающий с mock API.
const ConsentRequiredMessage = "Не подтверждено согласие на обработку персональных данных"

// Ограничения формы.
const (
	// MaxPickedIdeas — сколько идей можно приложить к записи.
	MaxPickedIdeas = 5
	// MaxFiles — сколько файлов дизайн-проекта принимается.
	MaxFiles = 10
	// MaxWishLen — длина пожеланий.
	MaxWishLen = 1200
)

// TaskOptions — задачи визита.
var TaskOptions = []string{"Подбор интерьера", "Мебель и материалы", "Ремонт", "Свой дизайн-проект"}

// TimeOptions — удобное время.
var TimeOptions = []string{"Уточнить по телефону", "Первая половина дня", "Вторая половина дня"}

// PlaceAddress — адрес Интерьерного центра.
const PlaceAddress = "Москва, Аминьевское шоссе, 6, ТЦ «Квартал», 2 этаж"

// phoneRe повторяет регулярное выражение mock API.
var phoneRe = regexp.MustCompile(`^[+0-9\s()\-]{10,22}$`)

// Request — заявка на запись.
type Request struct {
	Name          string
	Phone         string
	Tasks         []string
	PreferredDate string // YYYY-MM-DD или пусто
	PreferredTime string
	Ideas         []string
	Files         []string
	Wish          string
	Consent       bool
}

// Validation — результат проверки формы.
type Validation struct {
	OK       bool
	Errors   []string
	Warnings []string
	Summary  []string
}

// Appointment — сохранённая запись.
type Appointment struct {
	ID          string
	Status      string
	CreatedAt   time.Time
	Request     Request
	Validation  Validation
	Idempotent  bool
}

// Place — адрес и контакты центра для формы.
type Place struct {
	Address   string
	Mall      string
	Phone     string
	PhoneHref string
	Map       string
	Note      string
}

// Options — опции формы записи.
type Options struct {
	Tasks      []string
	Times      []string
	MaxIdeas   int
	MaxFiles   int
	Place      Place
	ConsentURL string
	PolicyURL  string
	DryRun     string
	Create     string
}

// DefaultOptions возвращает опции, дословно совпадающие с mock API.
func DefaultOptions() Options {
	return Options{
		Tasks:    TaskOptions,
		Times:    TimeOptions,
		MaxIdeas: MaxPickedIdeas,
		MaxFiles: MaxFiles,
		Place: Place{
			Address:   "Москва, Аминьевское шоссе, 6",
			Mall:      "ТЦ «Квартал», 2 этаж · бесплатная парковка",
			Phone:     "+7 (495) 138-73-38",
			PhoneHref: "tel:+74951387338",
			Map: "https://yandex.ru/maps/?text=%D0%9C%D0%BE%D1%81%D0%BA%D0%B2%D0%B0%2C%20" +
				"%D0%90%D0%BC%D0%B8%D0%BD%D1%8C%D0%B5%D0%B2%D1%81%D0%BA%D0%BE%D0%B5%20" +
				"%D1%88%D0%BE%D1%81%D1%81%D0%B5%2C%206%2C%20Zetta%20Home",
			Note: "Удобное время встречи подтвердит менеджер. Планируйте около 2–3 часов.",
		},
		ConsentURL: "https://zettahome.ru/privacy-policy/consent/",
		PolicyURL:  "https://zettahome.ru/privacy-policy/",
		DryRun:     "/api/center/appointment/dry-run",
		Create:     "/api/center/appointment",
	}
}

// Validate проверяет форму. Порядок ошибок совпадает с mock API.
func Validate(req Request, today time.Time) Validation {
	var errs, warns []string

	if strings.TrimSpace(req.Name) == "" {
		errs = append(errs, "Укажите имя")
	}
	if !phoneRe.MatchString(strings.TrimSpace(req.Phone)) {
		errs = append(errs, "Телефон должен содержать 10-22 символа: цифры, +, скобки, дефисы")
	}
	if !req.Consent {
		errs = append(errs, ConsentRequiredMessage)
	}
	if len(req.Tasks) == 0 {
		warns = append(warns, "Не выбрана ни одна задача — менеджер уточнит её на связи")
	}
	var unknown []string
	for _, t := range req.Tasks {
		if !contains(TaskOptions, t) {
			unknown = append(unknown, t)
		}
	}
	if len(unknown) > 0 {
		errs = append(errs, "Неизвестные задачи: "+strings.Join(unknown, ", "))
	}
	if !contains(TimeOptions, req.PreferredTime) {
		errs = append(errs, "Недопустимое удобное время")
	}
	if req.PreferredDate != "" {
		parsed, err := time.Parse("2006-01-02", req.PreferredDate)
		if err != nil {
			errs = append(errs, "Дата визита уже прошла")
		} else if parsed.Before(startOfDay(today)) {
			errs = append(errs, "Дата визита уже прошла")
		}
	}
	if len(req.Ideas) > MaxPickedIdeas {
		errs = append(errs, "Для подготовки подборки можно выбрать не более "+
			itoa(MaxPickedIdeas)+" идей")
	}
	if len(req.Files) > MaxFiles {
		errs = append(errs, "Не более "+itoa(MaxFiles)+" файлов дизайн-проекта")
	}

	summary := []string{"Задачи: " + orDefault(strings.Join(req.Tasks, ", "), "не выбраны")}
	if req.PreferredDate != "" {
		summary = append(summary, "Дата: "+req.PreferredDate+", "+req.PreferredTime)
	} else {
		summary = append(summary, "Время: "+req.PreferredTime)
	}
	if len(req.Ideas) > 0 {
		summary = append(summary, "Идей к подборке: "+itoa(len(req.Ideas)))
	}
	if len(req.Files) > 0 {
		summary = append(summary, "Файлов дизайн-проекта: "+itoa(len(req.Files)))
	}
	if strings.TrimSpace(req.Wish) != "" {
		summary = append(summary, "Пожелания заполнены")
	}

	if errs == nil {
		errs = []string{}
	}
	if warns == nil {
		warns = []string{}
	}
	return Validation{OK: len(errs) == 0, Errors: errs, Warnings: warns, Summary: summary}
}

// Repository — порт хранения записей.
type Repository interface {
	Save(ctx context.Context, a Appointment) error
	ByID(ctx context.Context, id string) (Appointment, error)
	Count(ctx context.Context) (int, error)
}

// Service — операции контекста Booking.
type Service struct {
	repo Repository
}

// NewService собирает сервис.
func NewService(repo Repository) *Service { return &Service{repo: repo} }

// DryRun проверяет форму без записи.
func (s *Service) DryRun(req Request, today time.Time) Validation { return Validate(req, today) }

// Create сохраняет запись. Вызывающий код обязан обеспечить idempotency и approval-gate.
func (s *Service) Create(ctx context.Context, a Appointment) error { return s.repo.Save(ctx, a) }

// ByID возвращает запись по идентификатору.
func (s *Service) ByID(ctx context.Context, id string) (Appointment, error) {
	return s.repo.ByID(ctx, id)
}

// Count возвращает число записей (для /api/health).
func (s *Service) Count(ctx context.Context) (int, error) { return s.repo.Count(ctx) }

func contains(list []string, v string) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}

func startOfDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func orDefault(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
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
