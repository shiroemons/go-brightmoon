// Package config はtitles_thコマンドの設定管理を行います
package config

import (
	"flag"
	"fmt"
	"os"
)

const Version = "0.0.3"

// Config はアプリケーションの設定を保持します
type Config struct {
	ArchivePath string
	ArchiveType int
	OutputDir   string
	DebugMode   bool
	DryRun      bool
	ShowVersion bool
}

// ParseFlags はコマンドライン引数を解析して設定を返します
func ParseFlags() *Config {
	config := &Config{}

	// カスタムUsage関数を設定（ダブルハイフン表示）
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		fmt.Fprintln(flag.CommandLine.Output(), "  --archive string")
		fmt.Fprintln(flag.CommandLine.Output(), "    \tpath to .dat archive file (e.g. th08.dat)")
		fmt.Fprintln(flag.CommandLine.Output(), "  -a string")
		fmt.Fprintln(flag.CommandLine.Output(), "    \tpath to .dat archive file (shorthand)")
		fmt.Fprintln(flag.CommandLine.Output(), "  --debug")
		fmt.Fprintln(flag.CommandLine.Output(), "    \tenable debug output")
		fmt.Fprintln(flag.CommandLine.Output(), "  -d\tenable debug output (shorthand)")
		fmt.Fprintln(flag.CommandLine.Output(), "  -o string")
		fmt.Fprintln(flag.CommandLine.Output(), "    \toutput directory for the generated files (default \".\")")
		fmt.Fprintln(flag.CommandLine.Output(), "  -t int")
		fmt.Fprintln(flag.CommandLine.Output(), "    \tarchive type (e.g., 0 for Imperishable Night, see README for details) (default -1)")
		fmt.Fprintln(flag.CommandLine.Output(), "  --dry-run")
		fmt.Fprintln(flag.CommandLine.Output(), "    \tperform a dry run without writing output files")
		fmt.Fprintln(flag.CommandLine.Output(), "  -n\tperform a dry run without writing output files (shorthand)")
		fmt.Fprintln(flag.CommandLine.Output(), "  --version")
		fmt.Fprintln(flag.CommandLine.Output(), "    \tshow version information")
		fmt.Fprintln(flag.CommandLine.Output(), "  -v\tshow version information (shorthand)")
	}

	// アーカイブフラグ
	flag.StringVar(&config.ArchivePath, "archive", "", "path to .dat archive file (e.g. th08.dat)")
	flag.StringVar(&config.ArchivePath, "a", "", "path to .dat archive file (e.g. th08.dat) (shorthand)")

	// タイプフラグ
	flag.IntVar(&config.ArchiveType, "t", -1, "archive type (e.g., 0 for Imperishable Night, see README for details)")

	// 出力ディレクトリ
	flag.StringVar(&config.OutputDir, "o", ".", "output directory for the generated files")

	// デバッグモード
	flag.BoolVar(&config.DebugMode, "debug", false, "enable debug output")
	flag.BoolVar(&config.DebugMode, "d", false, "enable debug output (shorthand)")

	// ドライランモード
	flag.BoolVar(&config.DryRun, "dry-run", false, "perform a dry run without writing output files")
	flag.BoolVar(&config.DryRun, "n", false, "perform a dry run without writing output files (shorthand)")

	// バージョン表示
	flag.BoolVar(&config.ShowVersion, "version", false, "show version information")
	flag.BoolVar(&config.ShowVersion, "v", false, "show version information (shorthand)")

	flag.Parse()

	return config
}

// HandleVersion はバージョン表示を処理します
func HandleVersion(showVersion bool) {
	if showVersion {
		fmt.Printf("titles_th version %s\n", Version)
		os.Exit(0)
	}
}

// DebugLogger はデバッグ出力を管理します
type DebugLogger struct {
	enabled bool
}

// NewDebugLogger は新しいDebugLoggerを作成します
func NewDebugLogger(enabled bool) *DebugLogger {
	return &DebugLogger{enabled: enabled}
}

// Printf はデバッグモードが有効な場合のみメッセージを表示します
func (d *DebugLogger) Printf(format string, a ...any) {
	if d.enabled {
		fmt.Printf(format, a...)
	}
}
