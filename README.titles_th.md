# 曲目ファイル作るくん (titles_th)

東方Projectのゲームから曲目情報（thbgm.fmt、musiccmt.txt）を抽出し、タイトルファイルを生成するツールです。

## 使い方

### 方法1: ダブルクリックで実行（Windows）

1. 東方Projectのゲームフォルダ（`thXX.dat` や `thbgm.dat` が存在するフォルダ）に、`titles_th.exe` を配置します
   - 体験版の場合は、`thXXtr.dat` や `thbgm_tr.dat` があるフォルダに配置してください

2. `titles_th.exe` をダブルクリックして実行します
   - セキュリティ警告が表示される場合がありますが、「実行」ボタンをクリックしてください
   - 「Microsoft Defender SmartScreen」などの警告は、未署名のプログラムに対する一般的な警告です

3. 自動的に処理が行われ、同じフォルダ内に `titles_thXX.txt` ファイルが生成されます
   - 体験版の場合は `titles_thXXtr.txt` が生成されます

### 方法2: コマンドラインで実行

```bash
# アーカイブから直接タイトル情報を抽出
titles_th -a th10.dat

# ドライランで動作確認（ファイル生成なし）
titles_th -a th10.dat --dry-run

# 出力先を指定
titles_th -a th10.dat -o output
```

詳細なオプションについては [README.md](README.md) を参照してください。

## 出力ファイル

生成される `titles_thXX.txt` には、各曲の以下の情報が記録されています：
- 開始位置
- イントロ部の長さ
- ループ部の長さ
- 曲名

出力ファイルは UTF-8 BOM 付きで保存されます。

## 注意事項

- このツールは、東方Projectのゲームアーカイブからデータを読み取るだけで、ファイルを改変することはありません
