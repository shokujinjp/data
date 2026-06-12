# AGENTS.md

食神（調布の中華料理店、X: [@shokujinjp](https://x.com/shokujinjp)）のオリジナルメニューデータを管理するリポジトリ。

## weekly.csv のフォーマット

```
id,name,price,category,day_start,day_end,can_weekday,description
2023071709,海鮮ミックスと豆腐の旨煮,650,定食,2023-07-17,2023-07-23,,週代わり定食9番
```

- **ヘッダ行なし**、1行目からデータ
- `id`: `YYYYMMDD` + `09` または `15`（週の月曜日 + メニュー番号）
- `day_start` は必ず月曜、`day_end` はその週の日曜（+6日）
- `description` は `週代わり定食9番` / `週代わり定食15番`
- 「週代わり」「週替わり」など表記は複数あり得るため、追記時はその時々の既存データ・対象箇所の表記に合わせる
- 価格はメニュー貼り紙の印字に準拠

## データソースと取得方法

週替わりメニューは毎週月曜の午前（JST）に X へ投稿される:

- 「今週の週替わり定食」ツイート: メニュー貼り紙の写真1枚（これが正典。`9.<メニュー名> <価格>円 / 15.<メニュー名> <価格>円` が印字されている）
- 「#食神週替わり定食」ツイート: 料理写真3枚（補助的に使用）

旧自動取得ツール `gen_weekly`（Twitter API + Cloud Vision OCR）は 2023-07 の API 有料化以降動作していなかったため 2026-06 に一度削除し、同月に LLM ベースで再実装した。現在の構成:

- `/gen-menu`（`.opencode/command/gen-menu.md`、Claude Code 用ラッパーは `.claude/commands/gen-menu.md`）: ツイート取得と画像読み取りの手順書。週替わり定食に加え、冷やし中華などの期間限定メニュー告知にも対応
- `gen_weekly/`: 読み取った内容を検証して `weekly.csv` / `limited.csv` に追記する Go CLI（重複チェック・月曜チェック付き）
- `.github/workflows/post.yaml`: 毎週月曜 13:00 JST に opencode で手順書を実行し、変更があれば PR を作成。`OPENCODE_API_KEY` secret（OpenCode Go / Zen、https://opencode.ai/auth）が必要。モデルはリポジトリ変数 `OPENCODE_MODEL` で変更可能（デフォルト `opencode/kimi-k2.6`。貼り紙画像を読むため vision 対応モデルであること）

2026-06 に 2023-07-17〜2026-06-08 分を手動バックフィル済み（PR #31）。

### バックフィルで得た知見（2026-06 時点）

- **X 本体**: 未ログインではプロフィールの人気ツイートのみ表示され、検索・過去遡りは不可。`cdn.syndication.twimg.com` / `syndication.twitter.com` のタイムライン API も人気ツイート約100件のみ
- **nitter インスタンス**: `nitter.tiekoetter.com` が検索（`/shokujinjp/search?f=tweets&q=...&since=...&until=...`）込みで動作する。Anubis（Proof-of-Work ボット対策）があり、curl・headless ブラウザは拒否されるが **headed ブラウザなら通過できる**（`npx agent-browser --headed`）。`r.jina.ai` も Anubis は通過不可
- **検索の取りこぼしに注意**: キーワード検索（`q=週替わり`）は普通に数週間分を取りこぼす。日付ウィンドウを5週間程度に区切って巡回し、欠落週は空クエリ（全件）検索を週単位で再実行して補完する。1ページ20件で打ち切られるため、リプライが多い週はウィンドウをさらに狭める
- **画像取得**: ツイート画像は `https://pbs.twimg.com/media/<MEDIA_ID>?format=jpg&name=medium` で認証なしに直接取得できる
- **読み取り**: 貼り紙は活字で鮮明なため、画像をマルチモーダルで読めば OCR 不要

## データ上の特記事項

- 価格遷移: 650/780円 → 700/830円（2024-11-04〜） → 750/880円（2025-04-07〜）
- 2026-01-05〜2026-03-16 は貼り紙に15番のみ記載（9番なし）。2026-03-23 から9番復活
- 盆・正月・夏季休業・臨時休業などでメニュー掲示が無い週は行を作らない（欠番が正常）
- 2025-06-16 週の9番は貼り紙印字どおり650円で記録（前後は750円のため誤植の可能性）。同週15番は価格未印字のため当時の標準価格880円を採用
- メニュー名は貼り紙の表記に準拠しており、表記ゆれ（えのき/エリンギ、インゲン/いんげん、豚バラ肉/豚肉 など）はそのまま残している

## 開発メモ

- コミットは Conventional Commits 形式、ブランチは `feat/<description>` / `fix/<description>` パターン
- GitHub Actions: `ci.yaml`（actionlint + gen_weekly の go vet / go test）と `post.yaml`（週次のメニュー自動取得）
