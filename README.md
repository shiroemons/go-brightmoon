# Brightmoon

Brightmoon は、東方 Project の `.dat` アーカイブを扱うための Go 製 CLI とライブラリです。
アーカイブ内ファイルの一覧表示、抽出、曲目ファイル生成用データの抽出に対応しています。

## 提供コマンド

- `brightmoon`: `.dat` アーカイブの一覧表示とファイル抽出を行う CLI
- `titles_th`（曲目ファイル作るくん）: `thbgm.fmt` と `musiccmt.txt` から `titles_*.txt` を生成する CLI

## 対応している主な処理

- アーカイブ形式の自動検出
- アーカイブ内エントリの一覧表示
- 全ファイルまたは指定ファイルのみの抽出
- 出力先ディレクトリ指定
- 並列抽出
- 曲目ファイル作るくんによる UTF-8 BOM 付き曲目ファイル生成
- Shift-JIS の `musiccmt.txt` / `readme.txt` 読み取り

## インストール

```bash
go install github.com/shiroemons/go-brightmoon/cmd/brightmoon@latest
go install github.com/shiroemons/go-brightmoon/cmd/titles_th@latest
```

## ソースからビルド

```bash
git clone https://github.com/shiroemons/go-brightmoon
cd go-brightmoon
make build
make build-brightmoon
make build-titles
```

個別に `go build` する場合:

```bash
go build -o brightmoon ./cmd/brightmoon
go build -o titles_th ./cmd/titles_th
```

## brightmoon

### 使い方

```text
brightmoon [オプション] <アーカイブファイル> [抽出ファイル1] [抽出ファイル2] ...
```

アーカイブファイル名の後にエントリ名を指定すると、そのファイルだけを抽出します。
抽出ファイルを指定せずに全ファイルを抽出する場合は `-x` を指定します。

### オプション

| オプション | 説明 | デフォルト |
|---|---|---|
| `-l` | アーカイブ内のファイル一覧を表示する | `false` |
| `-x` | アーカイブ内の全ファイルを抽出する | `false` |
| `-o <dir>` | 抽出先ディレクトリ | `.` |
| `-t <type>` | アーカイブサブタイプを明示指定する | `-1` |
| `-p` | 並列抽出を使う | `false` |
| `-w <num>` | 並列抽出のワーカー数 | `4` |
| `-d` | デバッグ情報を表示する | `false` |
| `-v` | バージョン情報を表示する | `false` |

### 使用例

```bash
# ファイル一覧を表示
brightmoon -l th10.dat

# 全ファイルを extracted/ に抽出
brightmoon -x -o extracted th08.dat

# 指定したエントリだけを抽出
brightmoon -o extracted th08.dat bgm/th08_01.wav

# 並列抽出
brightmoon -x -p -w 8 -o extracted th18.dat

# Kaguya Type 1 として明示的に開く
brightmoon -x -t 1 -o extracted th143.dat

# デバッグ情報を表示
brightmoon -d -l th07.dat

# バージョン表示
brightmoon -v
```

### アーカイブ自動検出

`-t` を省略した場合、実装は以下の流れで形式を判定します。

1. `Yukari`, `Yumemi`, `Suica`, `Hinanawi`, `Marisa`, `Kaguya`, `Kanako` の順に `Open` と `EnumFirst` を試す
2. 候補が 1 つならその形式を使う
3. 候補が複数ある場合は、ファイル名からゲーム番号を推測して形式を選ぶ
4. `Kaguya` または `Kanako` の場合は、ファイル名からサブタイプも推測する

ユーザーに選択を求めるプロンプトは表示されません。
判定できない場合やファイル名と候補が合わない場合はエラーになります。

### `-t` の現在の扱い

`-t` は実装上、主に `Kaguya` / `Kanako` のサブタイプ指定として扱われます。
現在の `openSpecificArchive` は `0` と `1` を先に `Kaguya` として解釈し、`2` を `Kanako` として解釈します。
そのため、TH10 から TH12.8 の Kanako Type 0 / Type 1 を手動で指定する用途には向きません。
通常はファイル名に基づく自動検出を使ってください。

| 値 | 現在の明示指定で開く形式 | 用途 |
|---|---|---|
| `0` | Kaguya Type 0 | 東方永夜抄 (TH08) |
| `1` | Kaguya Type 1 | 弾幕アマノジャク (TH14.3) |
| `2` | Kanako Type 2 | 東方神霊廟 (TH13) 以降 |

Kaguya の実装自体は Type 2 を東方花映塚 (TH09) 用として持っていますが、現在の CLI の明示指定では `-t 2` は Kanako Type 2 として解釈されます。

## 曲目ファイル作るくん (`titles_th`)

### 使い方

```text
titles_th [オプション]
```

曲目ファイル作るくんは `thbgm.fmt` と `musiccmt.txt` を読み取り、`titles_<入力名>.txt` を UTF-8 BOM 付きで生成します。
体験版ファイルの場合は `thbgm_tr.fmt` と `musiccmt_tr.txt` を読みます。

### 入力の探索順

1. `-a` / `--archive` が指定されている場合は、その `.dat` から対象ファイルを抽出する
2. 指定がない場合、カレントディレクトリから `thXX.dat` / `thXXtr.dat` を探す
3. カレントディレクトリになければ、実行ファイルと同じディレクトリから探す
4. `.dat` が見つからない場合、カレントディレクトリの `thbgm.fmt` / `musiccmt.txt` または `thbgm_tr.fmt` / `musiccmt_tr.txt` を読む

`thbgm.dat` は自動検出対象から除外されます。
同じ探索場所に対象の `.dat` が複数ある場合はエラーになります。

### オプション

| オプション | 説明 | デフォルト |
|---|---|---|
| `--archive`, `-a` | 入力 `.dat` アーカイブのパス | `""` |
| `-t <type>` | アーカイブサブタイプを明示指定する | `-1` |
| `-o <dir>` | 出力先ディレクトリ | `.` |
| `--debug`, `-d` | デバッグ情報を表示する | `false` |
| `--dry-run`, `-n` | ファイルを書き込まず、生成内容だけを標準出力に表示する | `false` |
| `--version`, `-v` | バージョン情報を表示する | `false` |

### 使用例

```bash
# カレントディレクトリの .dat またはローカルファイルから生成
titles_th

# アーカイブを明示指定
titles_th -a th19.dat

# 出力先を変更
titles_th -a th18.dat -o output

# ドライラン
titles_th -a th10.dat --dry-run

# デバッグ情報を表示
titles_th -a th20.dat -d

# バージョン表示
titles_th -v
```

ダブルクリックで実行する使い方は [README.titles_th.md](README.titles_th.md) を参照してください。

## 対応形式

公開パッケージ `pkg/pbgarc` には以下のアーカイブ実装があります。

| 実装 | 主な対象 | 備考 |
|---|---|---|
| `HinanawiArchive` | 東方紅魔郷 (TH06) | 旧形式 |
| `YukariArchive` | PBG4 形式 | `YukariMagic` は `PBG4` |
| `YumemiArchive` | 旧形式 | 8.3 形式のエントリ名を扱う |
| `KaguyaArchive` | TH08 / TH09 / TH14.3 系 | `SetArchiveType` で暗号パラメータを切り替える |
| `MarisaArchive` | 東方文花帖 (TH09.5) | 専用形式 |
| `KanakoArchive` | TH10 以降の THA1 形式 | `SetArchiveType` で暗号パラメータを切り替える |
| `SuicaArchive` | TH10 系の別形式 | 専用実装 |

ファイル名ベースの自動判定では、おおむね以下のように扱います。

| ファイル名例 | 想定形式 | 自動サブタイプ |
|---|---|---|
| `th06*.dat` | Hinanawi | - |
| `th07*.dat` | Yukari / Yumemi | - |
| `th08*.dat` | Kaguya | `0` |
| `th09*.dat` | Kaguya | `2` |
| `th095*.dat` | Marisa または Kanako Type 0 | コマンド側の判定経路に依存 |
| `th10*.dat`, `th11*.dat` | Kanako | `0` |
| `th12*.dat`, `th125*.dat`, `th128*.dat` | Kanako | `1` |
| `th13*.dat` 以降 | Kanako | `2` |

## ライブラリ構成

```text
go-brightmoon/
├── cmd/
│   ├── brightmoon/         # アーカイブ操作 CLI
│   └── titles_th/          # 曲目ファイル作るくん
├── pkg/
│   ├── pbgarc/             # アーカイブ形式の実装
│   └── crypto/             # 暗号化、復号、展開処理
└── internal/titles/
    ├── app/                # 曲目ファイル作るくんのアプリケーションロジック
    ├── archive/            # 曲目ファイル作るくん用アーカイブ抽出
    ├── config/             # CLI 設定
    ├── fileutil/           # ファイル探索、保存、文字コード変換
    ├── interfaces/         # テスト用インターフェース
    ├── models/             # データモデル
    └── parser/             # thbgm.fmt / musiccmt.txt / readme.txt パーサー
```

### `PBGArchive` インターフェース

各アーカイブ形式は `pkg/pbgarc.PBGArchive` を実装します。

- `Open(filename)` / `Close()`
- `EnumFirst()` / `EnumNext()`
- `GetEntryName()`
- `GetOriginalSize()` / `GetCompressedSize()`
- `GetEntry()`
- `Extract(w, callback, user)`
- `ExtractAll(callback, user)`

## 開発

```bash
# 全パッケージをビルド
make build

# テストを race detector 付きで実行
make test

# カバレッジを出力
make test-cover

# go fmt
make fmt

# go vet
make vet

# golangci-lint v2.7.1 を Docker で実行
make lint

# 依存関係を検証
make mod-verify
```

## 注意事項

- ゲームデータ、抽出結果、ローカル生成物はコミットしないでください。
- `brightmoon` の抽出処理は、アーカイブ内エントリ名のパストラバーサルや絶対パスを拒否します。
- 曲目ファイル作るくんは入力アーカイブや `thbgm.*` / `musiccmt*.txt` を読み取るだけで、ゲームファイルを書き換えません。
