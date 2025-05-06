#!/bin/bash

set -e

# コマンドライン引数を検証
if [ $# -lt 1 ]; then
  echo "使用法: $0 <version> [titles_th]"
  echo "例: $0 1.0.0       # brightmoon のリリース"
  echo "例: $0 1.0.0 titles_th  # titles_th のリリース"
  exit 1
fi

VERSION=$1
CONFIG_FILE=".goreleaser.yaml"
PREFIX="brightmoon"

# タイトル抽出ツールのリリースかどうかを確認
if [ "$2" = "titles_th" ]; then
  PREFIX="titles_th"
  CONFIG_FILE=".goreleaser.titles_th.yaml"
fi

TAG="v${VERSION}-${PREFIX}"

echo "リリースバージョン: $VERSION"
echo "タグ: $TAG"
echo "使用する設定ファイル: $CONFIG_FILE"
echo ""

# 現在のブランチを確認
CURRENT_BRANCH=$(git branch --show-current)
echo "現在のブランチ: $CURRENT_BRANCH"
echo ""

# mainブランチでない場合は確認
if [ "$CURRENT_BRANCH" != "main" ]; then
  read -p "現在 main ブランチではありません。続行しますか？ (y/N): " CONFIRM
  if [ "$CONFIRM" != "y" ] && [ "$CONFIRM" != "Y" ]; then
    echo "リリースを中止しました。"
    exit 1
  fi
fi

# リモートタグの確認
if git ls-remote --tags origin | grep -q "refs/tags/$TAG"; then
  echo "リモートに $TAG タグが既に存在します。"
  read -p "リモートのタグを削除しますか？ (y/N): " DELETE_CONFIRM
  if [ "$DELETE_CONFIRM" != "y" ] && [ "$DELETE_CONFIRM" != "Y" ]; then
    echo "リリースを中止しました。"
    exit 1
  fi
  
  echo "リモートのタグを削除中..."
  git push --delete origin "$TAG"
fi

# ローカルタグの確認と削除
if git tag | grep -q "$TAG"; then
  echo "ローカルに $TAG タグが既に存在します。削除します..."
  git tag -d "$TAG"
fi

# 新しいタグを作成
echo "新しいタグ $TAG を作成中..."
git tag "$TAG"

# タグをプッシュ
echo "タグをリモートにプッシュ中..."
git push origin "$TAG"

echo ""
echo "🎉 リリースプロセスを開始しました！"
echo "GitHub Actions によってリリースが自動的に作成されます。"
echo "進捗は以下のURLで確認できます："
echo "https://github.com/shiroemons/go-brightmoon/actions" 