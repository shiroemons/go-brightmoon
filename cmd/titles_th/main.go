package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/shiroemons/go-brightmoon/internal/titles/app"
	"github.com/shiroemons/go-brightmoon/internal/titles/config"
)

func main() {
	// コマンドライン引数の解析
	cfg := config.ParseFlags()

	// バージョン表示の処理
	config.HandleVersion(cfg.ShowVersion)

	// コンテキストの作成（キャンセル可能）
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// シグナルハンドリングの設定
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
	}()

	// アプリケーションの実行
	application := app.New(cfg)
	if err := application.Run(ctx); err != nil {
		// コンテキストキャンセルの場合は特別なメッセージ
		if err == context.Canceled {
			fmt.Fprintf(os.Stderr, "\n処理がキャンセルされました\n")
			os.Exit(130) // 128 + SIGINT(2)
		}
		fmt.Fprintf(os.Stderr, "エラー: %v\n", err)
		os.Exit(1)
	}
}
