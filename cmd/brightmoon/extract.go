package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/shiroemons/go-brightmoon/pkg/pbgarc"
)

// 抽出ジョブを表す構造体
type extractJob struct {
	entry   pbgarc.PBGArchiveEntry
	outPath string
}

// 並列抽出処理に使用するコンテキスト
type extractContext struct {
	archive pbgarc.PBGArchive
	outDir  string
	jobs    chan extractJob
	results chan extractResult
	wg      sync.WaitGroup
	mu      sync.Mutex // 出力用のミューテックス
}

// 抽出結果
type extractResult struct {
	entryName string
	success   bool
	err       error
}

// 並列処理で抽出を実行
func extractArchiveParallel(archive pbgarc.PBGArchive, outDir string, numWorkers int, filesToExtract []string) (successCount int, notFoundFiles []string, err error) {
	if numWorkers <= 0 {
		numWorkers = 4 // デフォルトのワーカー数
	}

	// 出力ディレクトリを作成
	if errMkdir := os.MkdirAll(outDir, 0755); errMkdir != nil {
		err = fmt.Errorf("出力ディレクトリを作成できません: %v", errMkdir)
		return
	}

	// filesToExtract が指定されている場合、Setに変換して高速ルックアップ
	extractSet := make(map[string]bool)
	if len(filesToExtract) > 0 {
		for _, f := range filesToExtract {
			extractSet[f] = true
		}
	}

	// 抽出コンテキストを初期化
	ctx := &extractContext{
		archive: archive,
		outDir:  outDir,
		jobs:    make(chan extractJob, numWorkers*2),
		results: make(chan extractResult, numWorkers*2),
	}

	// ワーカーを起動
	for i := 0; i < numWorkers; i++ {
		ctx.wg.Add(1)
		go extractWorker(ctx)
	}

	// 結果処理用のgoroutineを起動
	var resultErr error
	resultDone := make(chan struct{})
	go func() {
		for result := range ctx.results {
			if result.success {
				successCount++
				if *debugFlag {
					ctx.mu.Lock()
					fmt.Printf("成功: %s\n", result.entryName)
					ctx.mu.Unlock()
				}
			} else {
				ctx.mu.Lock()
				fmt.Fprintf(os.Stderr, "抽出に失敗しました: %s - %v\n", result.entryName, result.err)
				ctx.mu.Unlock()
				if resultErr == nil { // 最初のエラーを保持
					resultErr = fmt.Errorf("抽出エラー: %s (%v)", result.entryName, result.err)
				}
			}
		}
		close(resultDone)
	}()

	// 全ファイルを列挙してジョブを投入
	if !archive.EnumFirst() {
		close(ctx.jobs)
		ctx.wg.Wait()
		close(ctx.results)
		<-resultDone
		err = fmt.Errorf("アーカイブにファイルがありません")
		return // successCount=0, notFoundFiles=filesToExtract (if any), err
	}

	foundFilesInSet := make(map[string]bool)
	do := true
	for do {
		entryName := archive.GetEntryName()

		// 特定ファイル抽出が有効かチェック
		if len(extractSet) > 0 {
			if _, shouldExtract := extractSet[entryName]; !shouldExtract {
				do = archive.EnumNext()
				continue // スキップ
			}
			foundFilesInSet[entryName] = true // 抽出対象として見つかったことを記録
		}

		outPath := filepath.Join(outDir, entryName)

		// ディレクトリを作成 (エラーは無視しない方が良い)
		if dir := filepath.Dir(outPath); dir != "." {
			if errMkdir := os.MkdirAll(dir, 0755); errMkdir != nil {
				ctx.mu.Lock()
				fmt.Fprintf(os.Stderr, "ディレクトリを作成できません %s: %v\n", dir, errMkdir)
				ctx.mu.Unlock()
				// ここでエラーをresultErrに設定することも検討
			}
		}

		// ジョブをキューに追加
		entry := archive.GetEntry()
		ctx.jobs <- extractJob{
			entry:   entry,
			outPath: outPath,
		}

		do = archive.EnumNext()
	}

	// 全てのジョブが投入されたらチャネルを閉じる
	close(ctx.jobs)

	// 全てのワーカーが終了するのを待つ
	ctx.wg.Wait()
	close(ctx.results)

	// 結果処理goroutineの終了を待つ
	<-resultDone

	// 指定されたファイルが見つからなかったものをリストアップ
	if len(extractSet) > 0 {
		for file := range extractSet {
			if !foundFilesInSet[file] {
				notFoundFiles = append(notFoundFiles, file)
			}
		}
	}

	err = resultErr // 抽出中の最初のエラーを設定
	return
}

// 抽出ワーカー
func extractWorker(ctx *extractContext) {
	defer ctx.wg.Done()

	for job := range ctx.jobs {
		// 出力ファイルを開く
		outFile, err := os.Create(job.outPath)
		if err != nil {
			ctx.results <- extractResult{
				entryName: job.entry.GetEntryName(),
				success:   false,
				err:       err,
			}
			continue
		}

		// バッファ付きライターを使用
		writer := bufio.NewWriter(outFile)

		// 抽出
		success := job.entry.Extract(writer, nil, nil)
		writer.Flush()
		outFile.Close()

		if !success {
			os.Remove(job.outPath)
			ctx.results <- extractResult{
				entryName: job.entry.GetEntryName(),
				success:   false,
				err:       fmt.Errorf("extraction failed"),
			}
		} else {
			ctx.results <- extractResult{
				entryName: job.entry.GetEntryName(),
				success:   true,
			}
		}
	}
}

// 並列処理なしでアーカイブを抽出（既存のコードを移植）
func extractArchiveSequential(archive pbgarc.PBGArchive, outDir string, filesToExtract []string) (successCount int, notFoundFiles []string, err error) {
	// 出力ディレクトリを作成
	if errMkdir := os.MkdirAll(outDir, 0755); errMkdir != nil {
		err = fmt.Errorf("出力ディレクトリを作成できません: %v", errMkdir)
		return
	}

	// filesToExtract が指定されている場合、Setに変換して高速ルックアップ
	extractSet := make(map[string]bool)
	if len(filesToExtract) > 0 {
		for _, f := range filesToExtract {
			extractSet[f] = true
		}
	}

	if !archive.EnumFirst() {
		err = fmt.Errorf("アーカイブにファイルがありません")
		return // successCount=0, notFoundFiles=filesToExtract (if any), err
	}

	foundFilesInSet := make(map[string]bool)
	var firstError error
	do := true
	for do {
		entryName := archive.GetEntryName()

		// 特定ファイル抽出が有効かチェック
		if len(extractSet) > 0 {
			if _, shouldExtract := extractSet[entryName]; !shouldExtract {
				do = archive.EnumNext()
				continue // スキップ
			}
			foundFilesInSet[entryName] = true // 抽出対象として見つかったことを記録
		}

		outPath := filepath.Join(outDir, entryName)

		// ディレクトリを作成
		if dir := filepath.Dir(outPath); dir != "." {
			if errMkdir := os.MkdirAll(dir, 0755); errMkdir != nil {
				fmt.Fprintf(os.Stderr, "ディレクトリを作成できません %s: %v\n", dir, errMkdir)
				// エラーがあっても続行するが、最初のエラーは記録しておく
				if firstError == nil {
					firstError = fmt.Errorf("ディレクトリ作成エラー: %s", dir)
				}
			}
		}

		// 出力ファイルを開く
		outFile, errCreate := os.Create(outPath)
		if errCreate != nil {
			fmt.Fprintf(os.Stderr, "ファイルを作成できません %s: %v\n", outPath, errCreate)
			if firstError == nil {
				firstError = fmt.Errorf("ファイル作成エラー: %s", outPath)
			}
			do = archive.EnumNext()
			continue
		}

		// 抽出
		writer := bufio.NewWriter(outFile)
		success := archive.Extract(writer, callback, nil) // callbackを渡すように修正
		flushErr := writer.Flush()
		closeErr := outFile.Close()

		if !success {
			fmt.Fprintf(os.Stderr, "抽出に失敗しました: %s\n", entryName)
			os.Remove(outPath) // 失敗したらファイルを削除
			if firstError == nil {
				firstError = fmt.Errorf("抽出失敗: %s", entryName)
			}
		} else if flushErr != nil {
			fmt.Fprintf(os.Stderr, "ファイル書き込み(Flush)に失敗しました: %s - %v\n", outPath, flushErr)
			if firstError == nil {
				firstError = fmt.Errorf("Flush失敗: %s", outPath)
			}
		} else if closeErr != nil {
			fmt.Fprintf(os.Stderr, "ファイル書き込み(Close)に失敗しました: %s - %v\n", outPath, closeErr)
			if firstError == nil {
				firstError = fmt.Errorf("Close失敗: %s", outPath)
			}
		} else {
			successCount++
		}

		do = archive.EnumNext()
	}

	// 指定されたファイルが見つからなかったものをリストアップ
	if len(extractSet) > 0 {
		for file := range extractSet {
			if !foundFilesInSet[file] {
				notFoundFiles = append(notFoundFiles, file)
			}
		}
	}

	err = firstError // 処理中の最初のエラーを設定
	return
}
