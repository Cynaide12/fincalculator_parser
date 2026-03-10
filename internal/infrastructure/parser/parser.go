package parser

import (
	"encoding/json"
	"fincalparser/pkg/logger/sl"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/robfig/cron/v3"
)

type ParserResourceType string

var (
	Bskrt ParserResourceType = "bashkortostan"
	RU    ParserResourceType = ""
	Krym  ParserResourceType = "krym"
	Tatar ParserResourceType = "tatarstan"
)

type ParserResource struct {
	Type *ParserResourceType
	Year int
}

var dayNames = map[int]string{
	1: "ПН",
	2: "ВТ",
	3: "СР",
	4: "ЧТ",
	5: "ПТ",
	6: "СБ",
	7: "ВС",
}

const baseUrl = "https://fincalculator.ru/kalendar"

type Parser struct {
	resource ParserResource
	doc      *goquery.Document
	log      *slog.Logger
	dataDir  string
}

type CalendarDay struct {
	Year        int       `json:"year"`
	Month       string    `json:"month"`
	MonthNumber int8      `json:"month_number"`
	Day         int       `json:"day"` // день месяца
	DayName     string    `json:"day_name"`
	DayType     string    `json:"day_type"`    //'working' 'weekend' 'shortened'
	WeekNumber  int       `json:"week_number"` //номер недели в году(52/53/54 и тп)
	Date        time.Time `json:"date"`
}

func New(resource ParserResource, log *slog.Logger, dataDir string) (*Parser, error) {
	if resource.Year == 0 {
		resource.Year = time.Now().Year()
	}
	p := &Parser{
		resource: resource,
		log:      log,
		dataDir:  dataDir,
	}
	doc, err := p.getDocument()
	if err != nil {
		return nil, err
	}

	p.doc = doc

	return p, nil
}

func (p Parser) Start() {
	c := cron.New()
	c.AddFunc("@every 1d", p.execute)
	c.Start()
}

func (p Parser) execute() {
	data, err := p.getData()
	if err != nil {
		p.log.Error("ошибка при получении данных календаря", sl.Err(err))
	}
	if err := p.saveData(data); err != nil {
		p.log.Error("ошибка при сохранении данных", sl.Err(err))
	}
}

func (p Parser) ExecuteWithoutSaving() (*[]CalendarDay, error) {
	data, err := p.getData()
	return data, err
}

func (p Parser) getDocument() (*goquery.Document, error) {
	var url string
	if p.resource.Type == nil {
		url = fmt.Sprintf("%s/%d", baseUrl, p.resource.Year)
	} else if p.resource.Type != nil {
		url = fmt.Sprintf("%s/%d/%s", baseUrl, p.resource.Year, *p.resource.Type)
	}

	p.log.Info("делаю запрос")

	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	return doc, nil
}

func (p Parser) getData() (*[]CalendarDay, error) {

	p.log.Info("начинаю парсинг")

	calendarTables := p.doc.Find(".calendar.calendar__viewable > .row > .col-md-3.ng-star-inserted")

	timeLocation, err := time.LoadLocation("Asia/Yekaterinburg")
	if err != nil {
		return nil, err
	}

	var data []CalendarDay

	var e error

	calendarTables.Each(func(i int, s *goquery.Selection) {
		monthName := []rune(strings.TrimSpace(s.Find(".calendar_month-name").Text()))
		monthNumber := i + 1

		s.Find(".calendar-month-table_line.ng-star-inserted").Each(func(i int, z *goquery.Selection) {
			weekNumber, err := strconv.Atoi(z.Find(".calendar-month-table_week-number").Text())
			if err != nil {
				e = err
				return
			}
			z.Find(".calendar-month-day").Each(func(i int, d *goquery.Selection) {
				if d.Text() == "" {
					return
				}
				var dayType string
				if d.HasClass("calendar-month-day__dayoff") {
					dayType = "weekend"
				} else if d.HasClass("calendar-month-day__asterisk") {
					dayType = "shortened"
				} else {
					dayType = "working"
				}
				day, err := strconv.Atoi(d.Text())
				if err != nil {
					e = err
					return
				}
				data = append(data, CalendarDay{
					Year:        p.resource.Year,
					Month:       string(monthName[:3]),
					MonthNumber: int8(monthNumber),
					Day:         day,
					DayName:     dayNames[i+1],
					DayType:     dayType,
					WeekNumber:  weekNumber,
					Date:        time.Date(p.resource.Year, time.Month(monthNumber), day, 0, 0, 0, 0, timeLocation),
				})
			})

		})
	})
	if e != nil {
		return nil, e
	}

	p.log.Info("закончил парсинг")

	return &data, nil
}

func (p Parser) saveData(data *[]CalendarDay) error {
	err := os.MkdirAll(p.dataDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("не удалось создать директорию для хранения: %w", err)
	}

	file, err := os.Create(filepath.Join(p.dataDir, "data.json"))
	if err != nil {
		return fmt.Errorf("не удалось создать файл в папке назначения: %w", err)
	}
	defer file.Close()

	b, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("ошибка при маршаллинге данных в жсон: %w", err)
	}

	if _, err := file.Write(b); err != nil {
		return fmt.Errorf("не удалось сохранить данные: %w", err)
	}

	return nil
}

// получение дней с сохраненного жсон файла
func (p Parser) LoadData() ([]CalendarDay, error) {
	file, err := os.Open(filepath.Join(p.dataDir, "data.json"))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	b, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var data []CalendarDay
	if err := json.Unmarshal(b, &data); err != nil {
		return nil, err
	}

	return data, nil
}
