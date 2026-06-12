# gen_weekly

食神公式 X ([@shokujinjp](https://x.com/shokujinjp)) の告知から読み取ったメニュー情報を、検証付きで `weekly.csv` / `limited.csv` に追記する CLI。

旧 gen_weekly（Twitter API + Cloud Vision OCR）は 2023-07 の API 有料化で動作しなくなったため削除された。現在の構成では、ツイートの取得と画像の読み取りはコーディングエージェント（`/gen-menu` コマンド、`.opencode/command/gen-menu.md`。opencode と Claude Code の両方から使える）が行い、本 CLI は CSV への追記という決定的な処理だけを担当する。

## Usage

リポジトリルートから実行する。

```bash
# 週替わり定食（-date は必ずその週の月曜日。9番・15番は片方のみでも可）
go -C gen_weekly run . weekly -date 2026-06-15 \
  -name9 "豚肉と大根のピリ辛炒め" -price9 750 \
  -name15 "鶏肉と玉ねぎのクミン炒め" -price15 880

# 期間限定メニュー（冷やし中華など）
go -C gen_weekly run . limited -name "冷やし中華" -price 950 \
  -start 2026-06-01 -end 2026-09-30 -description "冷やし中華 2026"
```

- ID・カテゴリ・description は自動で付与する（weekly: `YYYYMMDD09`/`YYYYMMDD15`、limited: 開始日 + 同一開始日内の連番 2 桁）
- 登録済みデータは skip して正常終了するため、再実行は安全
- 月曜チェック・価格の数値チェック・期間の整合性チェックを行う

## 自動実行

`.github/workflows/post.yaml` が毎週月曜 13:00 JST に opencode（`opencode run`）で手順書を実行し、変更があれば PR を作成する。

- secret `OPENCODE_API_KEY`: OpenCode Go / Zen の API キー（https://opencode.ai/auth で発行）
- リポジトリ変数 `OPENCODE_MODEL`（任意）: 使用モデル。デフォルト `opencode/kimi-k2.6`。貼り紙画像を読むため vision 対応モデルを指定すること

## Test

```bash
go -C gen_weekly test ./...
```
