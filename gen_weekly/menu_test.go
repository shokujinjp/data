package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestWeeklyMenus(t *testing.T) {
	monday := time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)

	t.Run("9番と15番の両方", func(t *testing.T) {
		menus, err := weeklyMenus(monday, "豚肉と大根のピリ辛炒め", "750", "鶏肉と玉ねぎのクミン炒め", "880")
		if err != nil {
			t.Fatalf("weeklyMenus returned error: %v", err)
		}
		if len(menus) != 2 {
			t.Fatalf("len(menus) = %d, want 2", len(menus))
		}

		want9 := Menu{
			ID:          "2026061509",
			Name:        "豚肉と大根のピリ辛炒め",
			Price:       "750",
			Category:    "定食",
			DayStart:    "2026-06-15",
			DayEnd:      "2026-06-21",
			Description: "週代わり定食9番",
		}
		if menus[0] != want9 {
			t.Errorf("menus[0] = %+v, want %+v", menus[0], want9)
		}

		want15 := Menu{
			ID:          "2026061515",
			Name:        "鶏肉と玉ねぎのクミン炒め",
			Price:       "880",
			Category:    "定食",
			DayStart:    "2026-06-15",
			DayEnd:      "2026-06-21",
			Description: "週代わり定食15番",
		}
		if menus[1] != want15 {
			t.Errorf("menus[1] = %+v, want %+v", menus[1], want15)
		}
	})

	t.Run("15番のみ", func(t *testing.T) {
		menus, err := weeklyMenus(monday, "", "", "鶏肉と玉ねぎのクミン炒め", "880")
		if err != nil {
			t.Fatalf("weeklyMenus returned error: %v", err)
		}
		if len(menus) != 1 {
			t.Fatalf("len(menus) = %d, want 1", len(menus))
		}
		if menus[0].ID != "2026061515" {
			t.Errorf("menus[0].ID = %s, want 2026061515", menus[0].ID)
		}
	})

	t.Run("月曜以外はエラー", func(t *testing.T) {
		tuesday := monday.AddDate(0, 0, 1)
		_, err := weeklyMenus(tuesday, "麻婆豆腐", "750", "", "")
		if err == nil {
			t.Fatal("weeklyMenus accepted a non-Monday date")
		}
	})

	t.Run("メニュー未指定はエラー", func(t *testing.T) {
		_, err := weeklyMenus(monday, "", "", "", "")
		if err == nil {
			t.Fatal("weeklyMenus accepted empty menus")
		}
	})

	t.Run("名前と価格が揃っていなければエラー", func(t *testing.T) {
		_, err := weeklyMenus(monday, "麻婆豆腐", "", "", "")
		if err == nil {
			t.Fatal("weeklyMenus accepted a name without price")
		}
	})

	t.Run("価格が数値でなければエラー", func(t *testing.T) {
		_, err := weeklyMenus(monday, "麻婆豆腐", "750円", "", "")
		if err == nil {
			t.Fatal("weeklyMenus accepted a non-numeric price")
		}
	})
}

func TestLimitedMenu(t *testing.T) {
	existing := [][]string{
		{"id", "name", "price", "category", "day_start", "day_end", "can_weekday", "description"},
		{"2024060101", "冷やし中華", "900", "期間限定", "2024-06-01", "2024-09-30", "", "冷やし中華 2024"},
	}

	t.Run("新規メニュー", func(t *testing.T) {
		menu, err := limitedMenu(existing, "冷やし中華", "950", "2026-06-01", "2026-09-30", "冷やし中華 2026")
		if err != nil {
			t.Fatalf("limitedMenu returned error: %v", err)
		}
		want := Menu{
			ID:          "2026060101",
			Name:        "冷やし中華",
			Price:       "950",
			Category:    "期間限定",
			DayStart:    "2026-06-01",
			DayEnd:      "2026-09-30",
			Description: "冷やし中華 2026",
		}
		if menu != want {
			t.Errorf("menu = %+v, want %+v", menu, want)
		}
	})

	t.Run("同一開始日の連番", func(t *testing.T) {
		withSameDay := append(existing, []string{"2026060101", "冷やし中華", "950", "期間限定", "2026-06-01", "2026-09-30", "", "冷やし中華 2026"})
		menu, err := limitedMenu(withSameDay, "冷やし担々麺", "800", "2026-06-01", "2026-09-30", "冷やし担々麺 2026")
		if err != nil {
			t.Fatalf("limitedMenu returned error: %v", err)
		}
		if menu.ID != "2026060102" {
			t.Errorf("menu.ID = %s, want 2026060102", menu.ID)
		}
	})

	t.Run("同名同開始日は重複エラー", func(t *testing.T) {
		_, err := limitedMenu(existing, "冷やし中華", "900", "2024-06-01", "2024-09-30", "冷やし中華 2024")
		if !errors.Is(err, errAlreadyExists) {
			t.Fatalf("err = %v, want errAlreadyExists", err)
		}
	})

	t.Run("終了日が開始日より前はエラー", func(t *testing.T) {
		_, err := limitedMenu(existing, "冷やし中華", "950", "2026-09-30", "2026-06-01", "冷やし中華 2026")
		if err == nil {
			t.Fatal("limitedMenu accepted end before start")
		}
	})

	t.Run("日付形式が不正ならエラー", func(t *testing.T) {
		_, err := limitedMenu(existing, "冷やし中華", "950", "2026/06/01", "2026-09-30", "冷やし中華 2026")
		if err == nil {
			t.Fatal("limitedMenu accepted an invalid date format")
		}
	})
}

func TestAppendMenus(t *testing.T) {
	newFile := func(t *testing.T, content string) string {
		t.Helper()
		path := filepath.Join(t.TempDir(), "menu.csv")
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		return path
	}

	menu := Menu{
		ID:          "2026061509",
		Name:        "豚肉と大根のピリ辛炒め",
		Price:       "750",
		Category:    "定食",
		DayStart:    "2026-06-15",
		DayEnd:      "2026-06-21",
		Description: "週代わり定食9番",
	}

	t.Run("追記できる", func(t *testing.T) {
		path := newFile(t, "2026060809,豚肉とキャベツとちくわ炒め,750,定食,2026-06-08,2026-06-14,,週代わり定食9番\n")
		appended, err := appendMenus(path, []Menu{menu})
		if err != nil {
			t.Fatalf("appendMenus returned error: %v", err)
		}
		if len(appended) != 1 {
			t.Fatalf("len(appended) = %d, want 1", len(appended))
		}

		got, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		want := "2026060809,豚肉とキャベツとちくわ炒め,750,定食,2026-06-08,2026-06-14,,週代わり定食9番\n" +
			"2026061509,豚肉と大根のピリ辛炒め,750,定食,2026-06-15,2026-06-21,,週代わり定食9番\n"
		if string(got) != want {
			t.Errorf("file content = %q, want %q", got, want)
		}
	})

	t.Run("既存IDはスキップする", func(t *testing.T) {
		path := newFile(t, "2026061509,豚肉と大根のピリ辛炒め,750,定食,2026-06-15,2026-06-21,,週代わり定食9番\n")
		appended, err := appendMenus(path, []Menu{menu})
		if err != nil {
			t.Fatalf("appendMenus returned error: %v", err)
		}
		if len(appended) != 0 {
			t.Errorf("len(appended) = %d, want 0", len(appended))
		}
	})

	t.Run("末尾に改行がないファイルでも壊れない", func(t *testing.T) {
		path := newFile(t, "2026060809,豚肉とキャベツとちくわ炒め,750,定食,2026-06-08,2026-06-14,,週代わり定食9番")
		if _, err := appendMenus(path, []Menu{menu}); err != nil {
			t.Fatalf("appendMenus returned error: %v", err)
		}

		got, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		lines := strings.Split(strings.TrimRight(string(got), "\n"), "\n")
		if len(lines) != 2 {
			t.Errorf("got %d lines, want 2: %q", len(lines), got)
		}
	})

	t.Run("存在しないファイルはエラー", func(t *testing.T) {
		_, err := appendMenus(filepath.Join(t.TempDir(), "missing.csv"), []Menu{menu})
		if err == nil {
			t.Fatal("appendMenus accepted a missing file")
		}
	})
}
