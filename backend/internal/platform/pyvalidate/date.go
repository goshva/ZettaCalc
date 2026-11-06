package pyvalidate

import (
	"bytes"
	"math"
	"math/big"
	"strconv"
	"time"

	"github.com/zetta/zettacalc/backend/internal/platform/pyjson"
)

// Разбор даты повторяет pydantic v2 (speedate + pydantic-core), а не «правила
// ISO»: сообщения и приоритет проверок должны совпадать с mock API. Позиции
// считаются в байтах — speedate работает с байтами, а не с символами.

const (
	// msWatershed — порог speedate: |timestamp| больше него трактуется как
	// миллисекунды (2e10 секунд ≈ 2603 год, в миллисекундах ≈ 1970 год).
	msWatershed = int64(20_000_000_000)
	// unix0000 — 0000-01-01T00:00:00 как unix-timestamp.
	unix0000 = int64(-62_167_219_200)
	// unix9999 — 9999-12-31T23:59:59 как unix-timestamp.
	unix9999 = int64(253_402_300_799)
)

// DateResult — результат разбора поля `date | None`.
type DateResult struct {
	// Value — дата в формате YYYY-MM-DD (пусто, если значение null).
	Value string
	// Err — ошибка в формате pydantic, если значение недопустимо.
	Err *Error
}

// speedateDate — дата и время в терминах speedate.
type speedateDate struct {
	year, month, day     int
	hour, minute, second int
	micro                int
}

// DateValue разбирает значение поля `preferred_date` так, как это делает
// pydantic v2: строка как ISO-дата/датавремя или unix-timestamp, число как
// timestamp, всё остальное — date_type.
func DateValue(v *pyjson.Value) DateResult {
	if v == nil || v.Kind == pyjson.Null {
		return DateResult{}
	}
	switch v.Kind {
	case pyjson.String:
		return dateFromString(v.Str, v)
	case pyjson.Bool:
		return dateTypeError(v)
	case pyjson.Int:
		if n, ok := int64Value(v.IntValue); ok {
			return fromIntTimestamp(n, v)
		}
		// Значение не влезает в i64 — pydantic-core насыщает его.
		return fromIntTimestamp(saturateInt64(v.IntValue), v)
	case pyjson.Float:
		if math.IsNaN(v.FloatVal) {
			return speedateError("NaN values not permitted", v)
		}
		return fromFloatTimestamp(v.FloatVal, v)
	default:
		return dateTypeError(v)
	}
}

func dateTypeError(v *pyjson.Value) DateResult {
	return DateResult{Err: &Error{
		Type:  "date_type",
		Loc:   Body("preferred_date"),
		Msg:   "Input should be a valid date",
		Input: v,
	}}
}

// dateFromString повторяет Date::parse_bytes: сначала «чистая» дата или целое
// число как timestamp, затем датавремя или число с плавающей точкой.
//
// Итоговая ошибка берётся из разбора датавремя, как в DateTime::parse_bytes:
// именно её возвращает pydantic, когда ни один путь не сработал.
func dateFromString(s string, raw *pyjson.Value) DateResult {
	b := []byte(s)

	if dt, code := parseSpeedateDate(b); code == "" {
		return finishSpeedate(dt, raw)
	}
	if n, ok := intParseBytes(b); ok {
		return fromIntTimestamp(n, raw)
	}

	dt, code := parseSpeedateDateTime(b)
	if code == "" {
		return finishSpeedate(dt, raw)
	}
	if value, isFloat, ok := floatParseBytes(b); ok {
		if isFloat {
			return fromFloatTimestamp(value, raw)
		}
		return fromIntTimestamp(int64(value), raw)
	}
	return speedateError(code, raw)
}

// finishSpeedate проверяет «нулевое» время, затем год 0: порядок как в
// pydantic — сначала DateFromDatetimeInexact, затем DateParsing.
func finishSpeedate(dt speedateDate, raw *pyjson.Value) DateResult {
	if dt.hour != 0 || dt.minute != 0 || dt.second != 0 || dt.micro != 0 {
		return DateResult{Err: &Error{
			Type:  "date_from_datetime_inexact",
			Loc:   Body("preferred_date"),
			Msg:   "Datetimes provided to dates should have zero time - e.g. be exact dates",
			Input: raw,
		}}
	}
	if dt.year == 0 {
		return yearZeroError(raw)
	}
	return DateResult{Value: formatDate(dt.year, dt.month, dt.day)}
}

// fromFloatTimestamp повторяет DateTime::from_float_with_config: единица
// измерения выбирается по самому числу, доли округляются до микросекунд.
func fromFloatTimestamp(f float64, raw *pyjson.Value) DateResult {
	unit := int64(1000)
	scale := 1_000.0
	if math.Abs(f) <= float64(msWatershed) {
		unit = 1
		scale = 1_000_000.0
	}
	floor := math.Floor(f)
	micro := int(math.Round((f - floor) * scale))
	return fromTimestampWithUnit(saturateFloat(floor), micro, unit, raw)
}

// fromIntTimestamp повторяет Date::from_timestamp: единица измерения
// определяется порогом, время должно попадать ровно на полночь.
func fromIntTimestamp(n int64, raw *pyjson.Value) DateResult {
	unit := int64(1000)
	if abs64(n) <= msWatershed {
		unit = 1
	}
	return fromTimestampWithUnit(n, 0, unit, raw)
}

// fromTimestampWithUnit приводит timestamp к дате и проверяет точность.
func fromTimestampWithUnit(timestamp int64, micro int, unit int64, raw *pyjson.Value) DateResult {
	seconds := timestamp
	extra := micro
	if unit == 1000 {
		// timestamp_watershed: у отрицательных значений секунды и
		// микросекунды должны остаться одного знака.
		seconds = timestamp / 1000
		extra = int((timestamp%1000)*1000) + micro
		if extra < 0 {
			seconds--
			extra += 1_000_000
		}
	}
	if extra >= 1_000_000 {
		seconds += int64(extra / 1_000_000)
		extra %= 1_000_000
	}
	// Порядок как в speedate: сначала диапазон дат, затем точность.
	if seconds < unix0000 {
		return speedateError("dates before 0000 are not supported as unix timestamps", raw)
	}
	if seconds > unix9999 {
		return speedateError("dates after 9999 are not supported as unix timestamps", raw)
	}
	if seconds%86_400 != 0 || extra != 0 {
		return DateResult{Err: &Error{
			Type:  "date_from_datetime_inexact",
			Loc:   Body("preferred_date"),
			Msg:   "Datetimes provided to dates should have zero time - e.g. be exact dates",
			Input: raw,
		}}
	}
	t := time.Unix(seconds, 0).UTC()
	if t.Year() == 0 {
		return yearZeroError(raw)
	}
	return DateResult{Value: t.Format("2006-01-02")}
}

// parseSpeedateDate разбирает ровно дату (YYYY-MM-DD).
func parseSpeedateDate(b []byte) (speedateDate, string) {
	dt, code := parseSpeedateDatePartial(b)
	if code != "" {
		return speedateDate{}, code
	}
	if len(b) > 10 {
		return speedateDate{}, "unexpected extra characters at the end of the input"
	}
	return dt, ""
}

// parseSpeedateDateTime разбирает дату и время с необязательной таймзоной.
func parseSpeedateDateTime(b []byte) (speedateDate, string) {
	dt, code := parseSpeedateDatePartial(b)
	if code != "" {
		return speedateDate{}, code
	}
	if len(b) <= 10 {
		return speedateDate{}, "invalid datetime separator, expected `T`, `t`, `_` or space"
	}
	switch b[10] {
	case 'T', 't', ' ', '_':
	default:
		return speedateDate{}, "invalid datetime separator, expected `T`, `t`, `_` or space"
	}
	return parseSpeedateTime(b, 11, dt)
}

// parseSpeedateDatePartial повторяет Date::parse_bytes_partial: сначала
// символы, затем диапазоны месяца и дня.
func parseSpeedateDatePartial(b []byte) (speedateDate, string) {
	if len(b) < 10 {
		return speedateDate{}, "input is too short"
	}
	year, code := digitsAt(b, 0, 4, "invalid character in year")
	if code != "" {
		return speedateDate{}, code
	}
	if b[4] != '-' {
		return speedateDate{}, "invalid date separator, expected `-`"
	}
	month, code := digitsAt(b, 5, 2, "invalid character in month")
	if code != "" {
		return speedateDate{}, code
	}
	if b[7] != '-' {
		return speedateDate{}, "invalid date separator, expected `-`"
	}
	day, code := digitsAt(b, 8, 2, "invalid character in day")
	if code != "" {
		return speedateDate{}, code
	}
	if month < 1 || month > 12 {
		return speedateDate{}, "month value is outside expected range of 1-12"
	}
	if day < 1 || day > daysInMonth(year, month) {
		return speedateDate{}, "day value is outside expected range"
	}
	return speedateDate{year: year, month: month, day: day}, ""
}

// parseSpeedateTime повторяет Time::parse_bytes_offset: время, доли секунды и
// таймзона, позиции считаются в байтах.
func parseSpeedateTime(b []byte, offset int, dt speedateDate) (speedateDate, string) {
	if len(b)-offset < 5 {
		return speedateDate{}, "input is too short"
	}
	hour, code := digitsAt(b, offset, 2, "invalid character in hour")
	if code != "" {
		return speedateDate{}, code
	}
	if b[offset+2] != ':' {
		return speedateDate{}, "invalid time separator, expected `:`"
	}
	minute, code := digitsAt(b, offset+3, 2, "invalid character in minute")
	if code != "" {
		return speedateDate{}, code
	}
	dt.hour, dt.minute = hour, minute
	if dt.hour > 23 {
		return speedateDate{}, "hour value is outside expected range of 0-23"
	}
	if dt.minute > 59 {
		return speedateDate{}, "minute value is outside expected range of 0-59"
	}

	// Длина разобранной части: минимум HH:MM.
	length := 5
	if offset+5 < len(b) && b[offset+5] == ':' {
		second, code := digitsAt(b, offset+6, 2, "invalid character in second")
		if code != "" {
			return speedateDate{}, code
		}
		dt.second = second
		if dt.second > 59 {
			return speedateDate{}, "second value is outside expected range of 0-59"
		}
		length = 8
		if offset+8 < len(b) && (b[offset+8] == '.' || b[offset+8] == ',') {
			length = 9
			digits := 0
			value := 0
			for offset+length+digits < len(b) {
				c := b[offset+length+digits]
				if c < '0' || c > '9' {
					break
				}
				if digits < 6 {
					value = value*10 + int(c-'0')
				}
				digits++
			}
			if digits == 0 {
				return speedateDate{}, "second fraction digits missing after `.`"
			}
			if digits < 6 {
				for i := digits; i < 6; i++ {
					value *= 10
				}
			}
			dt.micro = value
			length += digits
		}
	}

	position := offset + length
	if position < len(b) {
		next := b[position]
		position++
		switch {
		case next == 'Z' || next == 'z':
			// Таймзона UTC: смещение не влияет на дату.
		case next == '+' || next == '-' || next == 0xE2:
			// U+2212 MINUS «−» разрешён ISO 8601 для отрицательных смещений.
			if next == 0xE2 {
				if position+1 >= len(b) || b[position] != 0x88 || b[position+1] != 0x92 {
					return speedateDate{}, "invalid timezone sign"
				}
				position += 2
			}
			tzHour, code := digitsAt(b, position, 2, "invalid timezone hour")
			if code != "" {
				return speedateDate{}, code
			}
			tzMinute := 0
			switch {
			case position+2 < len(b) && b[position+2] == ':':
				position += 3
				tzMinute, code = digitsAt(b, position, 2, "invalid timezone minute")
				if code != "" {
					return speedateDate{}, code
				}
			case position+2 < len(b) && b[position+2] >= '0' && b[position+2] <= '9':
				position += 2
				low, code := digitsAt(b, position+1, 1, "invalid timezone minute")
				if code != "" {
					return speedateDate{}, code
				}
				tzMinute = int(b[position]-'0')*10 + low
			default:
				return speedateDate{}, "invalid timezone minute"
			}
			if tzMinute >= 60 {
				return speedateDate{}, "timezone minute value is outside expected range of 0-59"
			}
			if tzHour*3600+tzMinute*60 >= 24*3600 {
				return speedateDate{}, "timezone offset must be less than 24 hours"
			}
			position += 2
		default:
			return speedateDate{}, "invalid timezone sign"
		}
	}
	if len(b) > position {
		return speedateDate{}, "unexpected extra characters at the end of the input"
	}
	return dt, ""
}

// digitsAt читает count цифр начиная с index.
func digitsAt(b []byte, index, count int, code string) (int, string) {
	value := 0
	for i := 0; i < count; i++ {
		if index+i >= len(b) {
			return 0, code
		}
		c := b[index+i]
		if c < '0' || c > '9' {
			return 0, code
		}
		value = value*10 + int(c-'0')
	}
	return value, ""
}

// intParseBytes повторяет int_parse_bytes из speedate.
func intParseBytes(b []byte) (int64, bool) {
	result := intParseBytesDiag(b)
	return result.value, result.err == intParseOK
}

// Виды сбоя целочисленного разбора speedate.
const (
	intParseOK intParseError = iota
	intParseEmpty
	intParseInvalidDigit
	intParseOverflow
)

// intParseError — причина сбоя int_parse_bytes_internal.
type intParseError int

// intParseResult — итог целочисленного разбора вместе с причиной сбоя.
type intParseResult struct {
	value int64
	err   intParseError
	// bad — первый недопустимый байт (для intParseInvalidDigit).
	bad byte
}

// intParseBytesDiag повторяет int_parse_bytes_internal: отдельно сообщает
// пустую строку, недопустимый символ и переполнение. Это важно для
// float_parse_bytes: к дробному разбору ведёт только точка в качестве первого
// недопустимого символа.
func intParseBytesDiag(b []byte) intParseResult {
	if len(b) == 0 {
		return intParseResult{err: intParseEmpty}
	}
	negative := false
	switch b[0] {
	case '-':
		negative = true
		b = b[1:]
	case '+':
		b = b[1:]
	}
	if len(b) == 0 {
		return intParseResult{err: intParseEmpty}
	}
	value := new(big.Int)
	for _, c := range b {
		if c < '0' || c > '9' {
			return intParseResult{err: intParseInvalidDigit, bad: c}
		}
		value.Mul(value, big.NewInt(10))
		value.Add(value, big.NewInt(int64(c-'0')))
	}
	// Ограничение i64::MAX действует для обоих знаков: i64::MIN строкой не
	// разбирается, как и в этой версии speedate.
	if value.Cmp(big.NewInt(math.MaxInt64)) > 0 {
		return intParseResult{err: intParseOverflow}
	}
	n := value.Int64()
	if negative {
		n = -n
	}
	return intParseResult{value: n, err: intParseOK}
}

// floatParseBytes повторяет float_parse_bytes: сначала целое, затем число с
// плавающей точкой. Дробный разбор включается, только если первым недопустимым
// символом целого оказалась точка (или точка есть при переполнении).
func floatParseBytes(b []byte) (float64, bool, bool) {
	result := intParseBytesDiag(b)
	switch result.err {
	case intParseOK:
		return float64(result.value), false, true
	case intParseInvalidDigit:
		if result.bad != '.' {
			return 0, false, false
		}
	case intParseOverflow:
		if !bytes.ContainsRune(b, '.') {
			return 0, false, false
		}
	default:
		return 0, false, false
	}
	if !looksLikeLexicalFloat(b) {
		return 0, false, false
	}
	f, err := strconv.ParseFloat(string(b), 64)
	if err != nil {
		return 0, false, false
	}
	return f, true, true
}

// looksLikeLexicalFloat проверяет формат числа: lexical STANDARD не принимает
// inf, nan, шестнадцатеричные записи и подчёркивания, но допускает «.5» и «5.».
func looksLikeLexicalFloat(b []byte) bool {
	i := 0
	if i < len(b) && (b[i] == '+' || b[i] == '-') {
		i++
	}
	before := 0
	for i < len(b) && b[i] >= '0' && b[i] <= '9' {
		i++
		before++
	}
	after := 0
	if i < len(b) && b[i] == '.' {
		i++
		for i < len(b) && b[i] >= '0' && b[i] <= '9' {
			i++
			after++
		}
	}
	if before == 0 && after == 0 {
		return false
	}
	if i < len(b) && (b[i] == 'e' || b[i] == 'E') {
		i++
		if i < len(b) && (b[i] == '+' || b[i] == '-') {
			i++
		}
		digits := 0
		for i < len(b) && b[i] >= '0' && b[i] <= '9' {
			i++
			digits++
		}
		if digits == 0 {
			return false
		}
	}
	return i == len(b)
}

// speedateError формирует ошибку разбора с текстом speedate.
func speedateError(code string, raw *pyjson.Value) DateResult {
	return DateResult{Err: &Error{
		Type:  "date_from_datetime_parsing",
		Loc:   Body("preferred_date"),
		Msg:   "Input should be a valid date or datetime, " + code,
		Input: raw,
		Ctx:   NewCtx().SetString("error", code),
	}}
}

func yearZeroError(raw *pyjson.Value) DateResult {
	return DateResult{Err: &Error{
		Type: "date_parsing",
		Loc:  Body("preferred_date"),
		Msg: "Input should be a valid date in the format YYYY-MM-DD, " +
			"year 0 is out of range",
		Input: raw,
		Ctx:   NewCtx().SetString("error", "year 0 is out of range"),
	}}
}

func formatDate(year, month, day int) string {
	return pad(year, 4) + "-" + pad(month, 2) + "-" + pad(day, 2)
}

func pad(n, width int) string {
	s := strconv.Itoa(n)
	for len(s) < width {
		s = "0" + s
	}
	return s
}

func daysInMonth(year, month int) int {
	switch month {
	case 1, 3, 5, 7, 8, 10, 12:
		return 31
	case 4, 6, 9, 11:
		return 30
	case 2:
		if (year%4 == 0 && year%100 != 0) || year%400 == 0 {
			return 29
		}
		return 28
	default:
		return 0
	}
}

func abs64(n int64) int64 {
	if n < 0 {
		return -n
	}
	return n
}

func int64Value(v *big.Int) (int64, bool) {
	if v.IsInt64() {
		return v.Int64(), true
	}
	return 0, false
}

func saturateInt64(v *big.Int) int64 {
	if v.Sign() < 0 {
		return math.MinInt64
	}
	return math.MaxInt64
}

func saturateFloat(f float64) int64 {
	switch {
	case math.IsNaN(f):
		return 0
	case f >= math.MaxInt64:
		return math.MaxInt64
	case f <= math.MinInt64:
		return math.MinInt64
	}
	return int64(f)
}
