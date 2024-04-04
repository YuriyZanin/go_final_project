package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Service struct {
	store Store
}

func NewService(store Store) Service {
	return Service{store: store}
}

func (s Service) Create(task Task) (Task, error) {
	id, err := s.store.Add(task)

	if err != nil {
		return task, err
	}

	task.Id = fmt.Sprint(id)
	return task, nil
}

func (s Service) FindBy(search string) ([]Task, error) {
	var res []Task
	var err error

	if search == "" {
		res, err = s.store.GetAll()
	} else if d, e := time.Parse("02.01.2006", search); e == nil {
		res, err = s.store.GetByDate(d)
	} else {
		res, err = s.store.GetByTitle(search)
	}

	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s Service) FindById(id int) (Task, error) {
	res, err := s.store.GetById(id)
	if err != nil {
		return res, err
	}

	return res, nil
}

func (s Service) Update(task Task) error {
	id, err := strconv.Atoi(task.Id)
	if err != nil {
		return err
	}
	_, err = s.FindById(id)
	if err != nil {
		return err
	}
	return s.store.Update(task)
}

func (s Service) Delete(id int) error {
	_, err := s.FindById(id)
	if err != nil {
		return err
	}

	return s.store.Delete(id)
}

func (s Service) NextDate(now time.Time, date string, repeat string) (string, error) {
	if repeat == "" {
		return "", fmt.Errorf("не указаны параметры")
	}

	startDate, err := time.Parse(DATE_FORMAT, date)
	if err != nil {
		return "", err
	}

	parts := strings.Split(repeat, " ")
	param := parts[0]
	switch param {
	case "y":
		currDate := startDate.AddDate(1, 0, 0)
		for now.After(currDate) || now.Equal(currDate) {
			currDate = currDate.AddDate(1, 0, 0)
		}

		return currDate.Format(DATE_FORMAT), nil

	case "d":
		if len(parts) == 1 {
			return "", fmt.Errorf("не указан интервал в днях")
		}

		days, err := strconv.Atoi(parts[1])
		if err != nil {
			return "", err
		}

		if days > 400 {
			return "", fmt.Errorf("превышен максимально допустимый интервал")
		}

		currDate := startDate.AddDate(0, 0, days)
		for now.After(currDate) {
			currDate = currDate.AddDate(0, 0, days)
		}

		return currDate.Format(DATE_FORMAT), nil

	case "w":
		if len(parts) == 1 {
			return "", fmt.Errorf("не указан интервал")
		}

		weekdays, err := extIntParams(parts[1], 1, 7)
		if err != nil {
			return "", err
		}

		currDate := now.AddDate(0, 0, 1)
		for {
			day := int(currDate.Weekday())
			for _, weekday := range weekdays {
				if day == weekday || (day == 0 && weekday == 7) {
					return currDate.Format(DATE_FORMAT), nil
				}
			}
			currDate = currDate.AddDate(0, 0, 1)
		}

	case "m":
		if len(parts) == 1 {
			return "", fmt.Errorf("не указан интервал")
		}

		monthDays, err := extIntParams(parts[1], -2, 31)
		if err != nil {
			return "", err
		}
		sortDayParams(monthDays)

		var months []int
		if len(parts) == 3 {
			months, err = extIntParams(parts[2], 1, 12)
			if err != nil {
				return "", err
			}
			sort.Ints(months)
		} else {
			s := int(startDate.Month())
			n := int(now.Month())
			if startDate.After(now) {
				months = []int{s, s + 1}
			} else {
				months = []int{n, n + 1}
			}
		}

		dateMap := buildDateMap(months)
		for _, d := range monthDays {
			for _, m := range months {
				days := dateMap[m]
				if len(days) <= d-1 {
					continue
				}

				var currDate time.Time
				if d < 0 {
					currDate = days[len(days)+d]
				} else {
					currDate = days[d-1]
				}

				if now.Before(currDate) || now.Equal(currDate) {
					return currDate.Format(DATE_FORMAT), nil
				}
			}
		}
		
		return "", fmt.Errorf("ошибка вычислния даты")

	default:
		return "", fmt.Errorf("неподдерживаемый формат %s", param)
	}
}

func buildDateMap(months []int) map[int][]time.Time {
	res := map[int][]time.Time{}
	for _, m := range months {
		currDay := time.Date(time.Now().Year(), time.Month(m), 1, 0, 0, 0, 0, time.UTC)
		lastDay := currDay.AddDate(0, 1, -1)
		res[m] = make([]time.Time, lastDay.Day())
		for i := 0; i < lastDay.Day(); i++ {
			res[m][i] = currDay
			currDay = currDay.AddDate(0, 0, 1)
		}
	}
	return res
}

func extIntParams(params string, min, max int) ([]int, error) {
	strings := strings.Split(params, ",")

	numbers := make([]int, len(strings))
	for i := 0; i < len(strings); i++ {
		number, err := strconv.Atoi(strings[i])
		if err != nil {
			return nil, err
		}
		if number < min || number > max {
			return nil, fmt.Errorf("недопустимое значение %d", number)
		}
		numbers[i] = number
	}

	return numbers, nil
}

func sortDayParams(days []int) {
	sort.SliceStable(days, func(i, j int) bool {
		if days[i] < 0 && days[j] < 0 {
			return days[i] < days[j]
		}
		if days[i] < 0 {
			return false
		}
		if days[j] < 0 {
			return true
		}
		return days[i] < days[j]
	})
}
