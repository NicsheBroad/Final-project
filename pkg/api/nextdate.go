package api

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const dateFormat = "20060102"

func afterNow(date, now time.Time) bool {
	return date.Format(dateFormat) > now.Format(dateFormat)
}

func startDate(now, date time.Time) time.Time {
	if afterNow(date, now) {
		return date
	}
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
}

func NextDate(now time.Time, dstart string, repeat string) (string, error) {
	if strings.TrimSpace(repeat) == "" {
		return "", errors.New("правило повторения не задано")
	}
	date, err := time.Parse(dateFormat, dstart)
	if err != nil {
		return "", fmt.Errorf("некорректная дата: %w", err)
	}

	parts := strings.Split(repeat, " ")
	switch parts[0] {
	case "d":
		return nextByDays(now, date, parts)
	case "y":
		if len(parts) != 1 {
			return "", errors.New("неверный формат правила y")
		}
		for {
			date = date.AddDate(1, 0, 0)
			if afterNow(date, now) {
				return date.Format(dateFormat), nil
			}
		}
	case "w":
		return nextByWeekdays(now, date, parts)
	case "m":
		return nextByMonthDays(now, date, parts)
	default:
		return "", errors.New("неподдерживаемый формат правила")
	}
}

func nextByDays(now, date time.Time, parts []string) (string, error) {
	if len(parts) != 2 {
		return "", errors.New("не указан интервал в днях")
	}
	n, err := strconv.Atoi(parts[1])
	if err != nil || n < 1 || n > 400 {
		return "", errors.New("недопустимый интервал в днях")
	}
	for {
		date = date.AddDate(0, 0, n)
		if afterNow(date, now) {
			return date.Format(dateFormat), nil
		}
	}
}

func nextByWeekdays(now, date time.Time, parts []string) (string, error) {
	if len(parts) != 2 {
		return "", errors.New("не указаны дни недели")
	}
	var days [8]bool
	for _, s := range strings.Split(parts[1], ",") {
		n, err := strconv.Atoi(s)
		if err != nil || n < 1 || n > 7 {
			return "", errors.New("недопустимый день недели")
		}
		days[n] = true
	}

	cur := startDate(now, date)
	for i := 0; i < 7; i++ {
		cur = cur.AddDate(0, 0, 1)
		wd := int(cur.Weekday())
		if wd == 0 {
			wd = 7
		}
		if days[wd] {
			return cur.Format(dateFormat), nil
		}
	}
	return "", errors.New("не удалось вычислить дату")
}

func nextByMonthDays(now, date time.Time, parts []string) (string, error) {
	if len(parts) < 2 || len(parts) > 3 {
		return "", errors.New("неверный формат правила m")
	}

	var days [32]bool
	var last, prevLast bool
	for _, s := range strings.Split(parts[1], ",") {
		n, err := strconv.Atoi(s)
		if err != nil {
			return "", errors.New("недопустимый день месяца")
		}
		switch {
		case n == -1:
			last = true
		case n == -2:
			prevLast = true
		case n >= 1 && n <= 31:
			days[n] = true
		default:
			return "", errors.New("недопустимый день месяца")
		}
	}

	var months [13]bool
	if len(parts) == 3 {
		for _, s := range strings.Split(parts[2], ",") {
			n, err := strconv.Atoi(s)
			if err != nil || n < 1 || n > 12 {
				return "", errors.New("недопустимый месяц")
			}
			months[n] = true
		}
	} else {
		for i := 1; i <= 12; i++ {
			months[i] = true
		}
	}

	cur := startDate(now, date)
	for i := 0; i < 366*8; i++ {
		cur = cur.AddDate(0, 0, 1)
		if !months[int(cur.Month())] {
			continue
		}
		lastDay := time.Date(cur.Year(), cur.Month()+1, 0, 0, 0, 0, 0, time.UTC).Day()
		d := cur.Day()
		if days[d] || (last && d == lastDay) || (prevLast && d == lastDay-1) {
			return cur.Format(dateFormat), nil
		}
	}
	return "", errors.New("не удалось вычислить дату")
}

func nextDateHandler(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	if s := r.FormValue("now"); s != "" {
		var err error
		now, err = time.Parse(dateFormat, s)
		if err != nil {
			http.Error(w, "некорректный параметр now", http.StatusBadRequest)
			return
		}
	}

	next, err := NextDate(now, r.FormValue("date"), r.FormValue("repeat"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
	w.Write([]byte(next))
}
