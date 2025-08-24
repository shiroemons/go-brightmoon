package main

import (
	"fmt"
	"os"

	"github.com/shiroemons/go-brightmoon/internal/titles/app"
	"github.com/shiroemons/go-brightmoon/internal/titles/config"
)

func main() {
	// コマンドライン引数の解析
	cfg := config.ParseFlags()

	// バージョン表示の処理
	config.HandleVersion(cfg.ShowVersion)

	// アプリケーションの実行
	application := app.New(cfg)
	if err := application.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "エラー: %v\n", err)
		os.Exit(1)
	}
}
