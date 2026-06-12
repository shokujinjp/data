package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	dayFormat = "2006-01-02"
	idFormat  = "20060102"

	weeklyDays = 6 // day_end は週の月曜 + 6日（日曜）

	categoryTeishoku = "定食"
	categoryLimited  = "期間限定"

	descWeekly9  = "週代わり定食9番"
	descWeekly15 = "週代わり定食15番"
)

var errAlreadyExists = errors.New("menu already exists")

// Menu は weekly.csv / limited.csv の1行を表す。
type Menu struct {
	ID          string
	Name        string
	Price       string
	Category    string
	DayStart    string
	DayEnd      string
	CanWeekday  string
	Description string
}

func (m Menu) record() []string {
	return []string{m.ID, m.Name, m.Price, m.Category, m.DayStart, m.DayEnd, m.CanWeekday, m.Description}
}

func validatePrice(price string) error {
	n, err := strconv.Atoi(price)
	if err != nil {
		return fmt.Errorf("price must be a number: %q", price)
	}
	if n <= 0 {
		return fmt.Errorf("price must be positive: %q", price)
	}
	return nil
}

// weeklyMenus は週替わり定食の告知内容から weekly.csv 向けの Menu を組み立てる。
// 9番・15番はどちらか一方のみでもよい（2026-01〜03 のように 15番のみの週がある）。
func weeklyMenus(monday time.Time, name9, price9, name15, price15 string) ([]Menu, error) {
	if monday.Weekday() != time.Monday {
		return nil, fmt.Errorf("date must be Monday: %s is %s", monday.Format(dayFormat), monday.Weekday())
	}

	dayStart := monday.Format(dayFormat)
	dayEnd := monday.AddDate(0, 0, weeklyDays).Format(dayFormat)

	menus := make([]Menu, 0, 2)
	for _, m := range []struct {
		number      string
		name        string
		price       string
		description string
	}{
		{"09", name9, price9, descWeekly9},
		{"15", name15, price15, descWeekly15},
	} {
		if m.name == "" && m.price == "" {
			continue
		}
		if m.name == "" || m.price == "" {
			return nil, fmt.Errorf("menu %s: name and price must be specified together", m.number)
		}
		if err := validatePrice(m.price); err != nil {
			return nil, fmt.Errorf("menu %s: %w", m.number, err)
		}
		menus = append(menus, Menu{
			ID:          monday.Format(idFormat) + m.number,
			Name:        m.name,
			Price:       m.price,
			Category:    categoryTeishoku,
			DayStart:    dayStart,
			DayEnd:      dayEnd,
			Description: m.description,
		})
	}

	if len(menus) == 0 {
		return nil, errors.New("no menu specified")
	}

	return menus, nil
}

// limitedMenu は期間限定メニュー（冷やし中華など）の告知内容から limited.csv 向けの
// Menu を組み立てる。ID は開始日 YYYYMMDD + 同一開始日内の連番2桁。
func limitedMenu(existing [][]string, name, price, start, end, description string) (Menu, error) {
	dayStart, err := time.Parse(dayFormat, start)
	if err != nil {
		return Menu{}, fmt.Errorf("invalid start date: %w", err)
	}
	dayEnd, err := time.Parse(dayFormat, end)
	if err != nil {
		return Menu{}, fmt.Errorf("invalid end date: %w", err)
	}
	if dayEnd.Before(dayStart) {
		return Menu{}, fmt.Errorf("end date %s is before start date %s", end, start)
	}
	if err := validatePrice(price); err != nil {
		return Menu{}, err
	}

	idPrefix := dayStart.Format(idFormat)
	seq := 1
	for _, record := range existing {
		if len(record) < 5 {
			continue
		}
		if record[1] == name && record[4] == start {
			return Menu{}, fmt.Errorf("%w: %s (day_start %s)", errAlreadyExists, name, start)
		}
		if !strings.HasPrefix(record[0], idPrefix) {
			continue
		}
		n, err := strconv.Atoi(strings.TrimPrefix(record[0], idPrefix))
		if err != nil {
			continue
		}
		if n >= seq {
			seq = n + 1
		}
	}

	return Menu{
		ID:          fmt.Sprintf("%s%02d", idPrefix, seq),
		Name:        name,
		Price:       price,
		Category:    categoryLimited,
		DayStart:    start,
		DayEnd:      end,
		Description: description,
	}, nil
}

func readRecords(path string) ([][]string, error) {
	fp, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fp.Close()

	return csv.NewReader(fp).ReadAll()
}

// appendMenus は menus のうち ID が未登録のものを path の CSV に追記し、
// 実際に追記した Menu を返す。
func appendMenus(path string, menus []Menu) ([]Menu, error) {
	records, err := readRecords(path)
	if err != nil {
		return nil, err
	}

	existingIDs := make(map[string]bool, len(records))
	for _, record := range records {
		if len(record) > 0 {
			existingIDs[record[0]] = true
		}
	}

	appended := make([]Menu, 0, len(menus))
	for _, menu := range menus {
		if !existingIDs[menu.ID] {
			appended = append(appended, menu)
		}
	}
	if len(appended) == 0 {
		return appended, nil
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	fp, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, err
	}
	defer fp.Close()

	if len(content) > 0 && content[len(content)-1] != '\n' {
		if _, err := fp.WriteString("\n"); err != nil {
			return nil, err
		}
	}

	writer := csv.NewWriter(fp)
	for _, menu := range appended {
		if err := writer.Write(menu.record()); err != nil {
			return nil, err
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, err
	}

	return appended, nil
}
