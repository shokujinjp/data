// gen_weekly は食神のメニュー CSV（weekly.csv / limited.csv）へ検証付きで
// 行を追記する CLI。ツイートの取得と読み取りは LLM（Claude Code）側が行い、
// 確定した内容を本 CLI に渡す。手順は .claude/commands/gen-menu.md を参照。
package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"time"
)

// gen_weekly ディレクトリ内から実行する前提（go -C gen_weekly run . ...）
const (
	defaultWeeklyFile  = "../weekly.csv"
	defaultLimitedFile = "../limited.csv"
)

func main() {
	log.SetFlags(0)

	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	var err error
	switch os.Args[1] {
	case "weekly":
		err = runWeekly(os.Args[2:])
	case "limited":
		err = runLimited(os.Args[2:])
	default:
		usage()
		os.Exit(2)
	}
	if err != nil {
		log.Fatalf("error: %s", err)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, `usage:
  gen_weekly weekly  -date YYYY-MM-DD [-name9 NAME -price9 PRICE] [-name15 NAME -price15 PRICE] [-file weekly.csv]
  gen_weekly limited -name NAME -price PRICE -start YYYY-MM-DD -end YYYY-MM-DD -description DESC [-file limited.csv]`)
}

func runWeekly(args []string) error {
	fs := flag.NewFlagSet("weekly", flag.ExitOnError)
	date := fs.String("date", "", "週の月曜日 (YYYY-MM-DD)")
	name9 := fs.String("name9", "", "9番のメニュー名")
	price9 := fs.String("price9", "", "9番の価格")
	name15 := fs.String("name15", "", "15番のメニュー名")
	price15 := fs.String("price15", "", "15番の価格")
	file := fs.String("file", defaultWeeklyFile, "weekly.csv のパス")
	if err := fs.Parse(args); err != nil {
		return err
	}

	monday, err := time.Parse(dayFormat, *date)
	if err != nil {
		return fmt.Errorf("invalid -date: %w", err)
	}

	menus, err := weeklyMenus(monday, *name9, *price9, *name15, *price15)
	if err != nil {
		return err
	}

	appended, err := appendMenus(*file, menus)
	if err != nil {
		return err
	}

	reportAppended(*file, menus, appended)
	return nil
}

func runLimited(args []string) error {
	fs := flag.NewFlagSet("limited", flag.ExitOnError)
	name := fs.String("name", "", "メニュー名")
	price := fs.String("price", "", "価格")
	start := fs.String("start", "", "提供開始日 (YYYY-MM-DD)")
	end := fs.String("end", "", "提供終了日 (YYYY-MM-DD)")
	description := fs.String("description", "", "説明（例: 冷やし中華 2026）")
	file := fs.String("file", defaultLimitedFile, "limited.csv のパス")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *name == "" || *description == "" {
		return errors.New("-name and -description are required")
	}

	records, err := readRecords(*file)
	if err != nil {
		return err
	}

	menu, err := limitedMenu(records, *name, *price, *start, *end, *description)
	if errors.Is(err, errAlreadyExists) {
		log.Printf("skip: %s", err)
		return nil
	}
	if err != nil {
		return err
	}

	appended, err := appendMenus(*file, []Menu{menu})
	if err != nil {
		return err
	}

	reportAppended(*file, []Menu{menu}, appended)
	return nil
}

func reportAppended(file string, requested, appended []Menu) {
	appendedIDs := make(map[string]bool, len(appended))
	for _, menu := range appended {
		appendedIDs[menu.ID] = true
	}
	for _, menu := range requested {
		if appendedIDs[menu.ID] {
			log.Printf("appended to %s: %s (%s)", file, menu.ID, menu.Name)
		} else {
			log.Printf("skip: %s already exists in %s", menu.ID, file)
		}
	}
}
