package main

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

func parseRelativeTime(input string) (time.Time, bool) {
	now := time.Now()
	relativeTimePattern := regexp.MustCompile(`^(\d+)\s*(秒前|分钟前|小时前|天前|周前|月前|年前)$`)
	matches := relativeTimePattern.FindStringSubmatch(input)

	if len(matches) != 3 {
		return time.Time{}, false
	}

	value, err := strconv.Atoi(matches[1])
	if err != nil {
		return time.Time{}, false
	}

	var duration time.Duration
	switch matches[2] {
	case "秒前":
		duration = time.Duration(value) * time.Second
	case "分钟前":
		duration = time.Duration(value) * time.Minute
	case "小时前":
		duration = time.Duration(value) * time.Hour
	case "天前":
		duration = time.Duration(value) * 24 * time.Hour
	case "周前":
		duration = time.Duration(value) * 7 * 24 * time.Hour
	case "月前":
		duration = time.Duration(value) * 30 * 24 * time.Hour
	case "年前":
		duration = time.Duration(value) * 365 * 24 * time.Hour
	default:
		return time.Time{}, false
	}

	return now.Add(-duration), true
}

func parseDate(input string, includeHour bool) (string, error) {
	dateLayouts := []string{
		"2006-01-02",
		"2006/01/02",
		"20060102",
		"2006.01.02",
		"01-02-2006",
		"01.02",
		"01/02/2006",
		"2006年01月02日",
		"01月02日",
		"2006年01月",
		"2 January, 2006",
		"January 2, 2006",
	}

	dateTimeLayouts := []string{
		"2006-01-02 15:04:05",
		"2006/01/02 15:04:05",
		"2006.01.02 15:04:05",
		"01-02-2006 15:04:05",
		"01/02/2006 15:04:05",
		"2006年01月02日 15:04:05",
		"2 January, 2006 15:04:05",
		"January 2, 2006 15:04:05",
		"2006-01-02T15:04:05.000Z",
	}

	layouts := dateLayouts
	if includeHour {
		layouts = append(layouts, dateTimeLayouts...)
	} else {
		// Attempt to parse the input as a date-time, but only return the date part
		for _, layout := range dateTimeLayouts {
			t, err := time.Parse(layout, input)
			if err == nil {
				return t.Format("2006-01-02"), nil
			}
		}
	}

	for _, layout := range layouts {
		t, err := time.Parse(layout, input)
		if err == nil {
			if includeHour {
				return t.Format("2006-01-02 15:04:05"), nil
			}
			return t.Format("2006-01-02"), nil
		}
	}

	return "", fmt.Errorf("未能解析日期: %s", input)
}

func dateFormatHandler(w http.ResponseWriter, r *http.Request) {
	dateParam := r.URL.Query().Get("date")
	hourParam := r.URL.Query().Get("hour")

	if dateParam == "" {
		http.Error(w, "缺少日期参数", http.StatusBadRequest)
		return
	}

	includeHour := hourParam == "true"
	relativeTime, ok := parseRelativeTime(dateParam)
	if ok {
		if includeHour {
			fmt.Fprint(w, relativeTime.Format("2006-01-02 15:04:05"))
		} else {
			fmt.Fprint(w, relativeTime.Format("2006-01-02"))
		}
		return
	}

	formattedDate, err := parseDate(dateParam, includeHour)
	if err != nil {
		http.Error(w, fmt.Sprintf("日期格式错误: %s", err.Error()), http.StatusBadRequest)
		return
	}

	fmt.Fprint(w, formattedDate)
}

func main() {
	http.HandleFunc("/convert", dateFormatHandler)
	fmt.Println("服务器已启动，监听 :8080 端口")
	http.ListenAndServe(":8080", nil)
}
