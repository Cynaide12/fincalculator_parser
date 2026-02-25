package parser

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"log/slog"
	"net/http"
)

type ParserResourceType string

var (
	Bskrt ParserResourceType = "bashkortostan"
	RU    ParserResourceType = ""
	Krym  ParserResourceType = "krym"
	Tatar ParserResourceType = "tatarstan"
)

const baseUrl = "https://fincalculator.ru/kalendar"

type Parser struct {
	resource ParserResourceType
	doc      *goquery.Document
	log      *slog.Logger
}

type CalendarDay struct {
	ID         int64
	Year       int
	Month      int
	Day        int
	IsWorking  bool
	IsHoliday  bool
	DayType    string //'working' 'weekend' 'holiday' 'shortened'
	WeekNumber int    //номер недели в году(52/53/54 и тп)
}

func NewParser(resource ParserResourceType, log *slog.Logger) (*Parser, error) {
	p := &Parser{
		resource: resource,
		log:      log,
	}
	doc, err := p.getDocument()
	if err != nil {
		return nil, err
	}

	p.doc = doc

	return p, nil
}

func (p Parser) getDocument() (*goquery.Document, error) {
	url := fmt.Sprintf("%s/%s", baseUrl, p.resource)

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

func (p Parser) GetDays() (*[]CalendarDay, error) {

	p.log.Info("начинаю парсинг")

	calendarTables := p.doc.Find(".calendar.calendar__viewable fc-calendar-month-table > .calendar-month-table")

	p.log.Debug(calendarTables.Text())

	// calendarTables.Find()

	return nil, nil
}
