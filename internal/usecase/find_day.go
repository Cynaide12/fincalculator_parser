package usecase

import (
	"fincalparser/internal/infrastructure/parser"
	"log/slog"
	"time"
)

type FindType string

var (
	Prev FindType = "prev"
	Next FindType = "next"
)

type FindDayUseCase struct {
	log    *slog.Logger
	parser *parser.Parser
}

func NewFindDayUseCase(log *slog.Logger, parser *parser.Parser) *FindDayUseCase {
	return &FindDayUseCase{log, parser}
}

func (uc *FindDayUseCase) Execute(startDate time.Time, daysPeriod int16, searchType FindType) (*parser.CalendarDay, error) {
	data, err := uc.parser.LoadData()
	if err != nil {
		return nil, err
	}

	timeLocation, err := time.LoadLocation("Asia/Yekaterinburg")
	if err != nil {
		return nil, err
	}

	if searchType == Next {
		return uc.findNextDay(data, startDate, daysPeriod, timeLocation)
	}
	return uc.findPrevDay(data, startDate, daysPeriod, timeLocation)
}

func (uc *FindDayUseCase) findNextDay(data []parser.CalendarDay, startDate time.Time, daysPeriod int16, loc *time.Location) (*parser.CalendarDay, error) {
	var checkedDays int16 = 0
	currentYear := startDate.Year()
	
	for {
		for _, day := range data {
			date := time.Date(currentYear, time.Month(day.MonthNumber), day.Day, 0, 0, 0, 0, loc)
			
			if date.After(startDate) {
				if checkedDays >= daysPeriod && day.DayType == "working" {
					day.Date = date
					day.Year = currentYear
					return &day, nil
				}
				checkedDays++
			}
		}
		currentYear++
	}
}

func (uc *FindDayUseCase) findPrevDay(data []parser.CalendarDay, startDate time.Time, daysPeriod int16, loc *time.Location) (*parser.CalendarDay, error) {
	var checkedDays int16 = 0
	currentYear := startDate.Year()
	
	for {
		for i := len(data) - 1; i >= 0; i-- {
			day := data[i]
			date := time.Date(currentYear, time.Month(day.MonthNumber), day.Day, 0, 0, 0, 0, loc)
			
			if date.Before(startDate) {
				if checkedDays >= daysPeriod && day.DayType == "working" {
					day.Date = date
					day.Year = currentYear
					return &day, nil
				}
				checkedDays++
			}
		}
		currentYear--
	}
}