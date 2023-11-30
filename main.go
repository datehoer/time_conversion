package main

import (
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
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

func parseSpecialDate(input string) (time.Time, error) {

	// Using a regular expression to match the special date formats (MM.DD or M月D日)
	specialDatePattern := regexp.MustCompile(`^(\d{1,2})[.月](\d{1,2})[日]*$`)
	matches := specialDatePattern.FindStringSubmatch(input)
	fmt.Printf("Attempting to parse special date: %s\n", input)
	if len(matches) != 3 {
		return time.Time{}, fmt.Errorf("Invalid special date format")
	}

	// Extracting the month and day
	month, err := strconv.Atoi(matches[1])
	if err != nil || month < 1 || month > 12 {
		return time.Time{}, fmt.Errorf("Invalid month value")
	}

	day, err := strconv.Atoi(matches[2])
	if err != nil || day < 1 || day > 31 {
		return time.Time{}, fmt.Errorf("Invalid day value")
	}

	// Using the current year
	currentYear := time.Now().Year()

	// Creating a time.Time object with the extracted values
	parsedDate := time.Date(currentYear, time.Month(month), day, 0, 0, 0, 0, time.UTC)

	return parsedDate, nil
}

func parseDate(input string, includeHour bool) (string, error) {

	//re := regexp.MustCompile(`^(.+)\s([\+\-]\d{2}:\d{2})$`)
	//matches := re.FindStringSubmatch(input)
	//if len(matches) == 3 {
	//	datetime, err := time.Parse("2006-01-02 15:04:05", matches[1])
	//	if err == nil {
	//		offset, err := time.ParseDuration(matches[2] + "0m")
	//		if err == nil {
	//			tz := time.FixedZone("UTC", int(offset.Seconds()))
	//			datetime = datetime.In(tz)
	//			if includeHour {
	//				return datetime.Format("2006-01-02 15:04:05"), nil
	//			}
	//			return datetime.Format("2006-01-02"), nil
	//		}
	//	}
	//}

	specialDate, err := parseSpecialDate(input)
	if err == nil {
		if includeHour {
			return specialDate.Format("2006-01-02 15:04:05"), nil
		}
		return specialDate.Format("2006-01-02"), nil
	}
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
		"2006-01-02 15:04:05 -07:00",
		"2006-01-02 15:04:05 -0700",
		"01/02/2006 15:04:05",
		"2006年01月02日 15:04:05",
		"2 January, 2006 15:04:05",
		"January 2, 2006 15:04:05",
		"2006-01-02T15:04:05.000Z",
		"2006-01-02T15:04:05Z",
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

func convertUsageHandler(c *gin.Context) {
	usage := `
    使用 /convert API 以将日期从一种格式转换为另一种格式。
    用法：
        GET /convert?date=DATE&hour=BOOLEAN
    参数：
        date - 要转换的日期。
        hour - 是否包括小时（可选，默认为false）。

    示例：
        GET /convert?date=01月02日&hour=true
		return 2023-01-02 00:00:00

		GET /convert?date=01月02日
		return 2023-01-02

		GET /convert?date=2 January, 2006
		return 2006-01-02
    `
	c.String(http.StatusOK, usage)
}

func main() {
	r := gin.Default()
	r.Use(cors.Default())
	r.GET("/convert", func(c *gin.Context) {
		w := c.Writer
		r := c.Request
		dateFormatHandler(w, r)
	})
	r.GET("/api", convertUsageHandler)
	// Define your routes and handlers here
	r.Run(":4447") // listen and serve on 0.0.0.0:4447 (for windows "localhost:4447")
}
