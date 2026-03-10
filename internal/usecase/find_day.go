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

func (uc *FindDayUseCase) Execute(startDate time.Time, daysPeriod int16) (*parser.CalendarDay, error) {
	var i int16 //счетчик проверенных дней
	var day *parser.CalendarDay = nil
	var year = time.Now().Year() //год найденного дня

	timeLocation, err := time.LoadLocation("Asia/Yekaterinburg")
	if err != nil {
		return nil, err
	}

	data, err := uc.parser.LoadData()
	if err != nil {
		return nil, err
	}
	for day == nil {
		for _, date := range data {
			//так как цикл может идти множество раз по повторяющемуся календарю(на следующий год данных не может быть)
			//то нужно дату у проверяемого дня увеличивать в соответствии с циклом
			date.Date = time.Date(year, time.Month(date.MonthNumber), date.Day, 0, 0, 0, 0, timeLocation)
			//если дата дня в шаге цикла > startDate
			if date.Date.Compare(startDate) > 0 {
				//если счетчик проверенных дней > периода даты И день является рабочим
				if i >= daysPeriod && date.DayType == "working" {
					date.Year = year
					uc.log.Debug("первое найденное", year, date.Date, startDate, date.MonthNumber)
					day = &date
					break
				}
				//иначе увеличиваем счетчик проверенных дней
				i++
			}
		}
		year++
	}

	return day, nil
}
