package usecase

import (
	"fincalparser/internal/infrastructure/parser"
	"log/slog"
	"time"
)

type FindDayUseCase struct {
	log    *slog.Logger
	parser *parser.Parser
}

func NewFindDayUseCase(log *slog.Logger, parser *parser.Parser) *FindDayUseCase {
	return &FindDayUseCase{log, parser}
}

//TODO: подумать че делать если startDate + daysPeriod указывает на следующий год. идти по циклу data по новой???
func (uc *FindDayUseCase) Execute(startDate time.Time, daysPeriod int16) (*parser.CalendarDay, error) {

	// обрезаем данные, оставляем только данные где дата >= startDate и где dayType == 'working'
	var slicedDates []parser.CalendarDay
	// year := startDate.Year()
	for true {
	if len(slicedDates) < int(daysPeriod) {
			data, err := uc.parser.LoadData()
			if err != nil {
				return nil, err
			}
			for _, date := range data {
				if startDate.Compare(date.Date) > 0 && date.DayType == "working" {
					slicedDates = append(slicedDates, date)
				}
			}
		}
		break
	}

	// var day *parser.CalendarDay = nil

	// while day == nil {
	// 	if len(slicedDates) > int(daysPeriod)
	// }

	return nil, nil
}
