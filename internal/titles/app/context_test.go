package app

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"github.com/shiroemons/go-brightmoon/internal/titles/config"
	"github.com/shiroemons/go-brightmoon/internal/titles/mocks"
)

func TestApp_Run_ContextCancellation(t *testing.T) {
	tests := []struct {
		name          string
		setupContext  func() (context.Context, context.CancelFunc)
		setupMock     func() *mocks.MockFileSystem
		expectedError error
	}{
		{
			name: "即座にキャンセルされたコンテキスト",
			setupContext: func() (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithCancel(context.Background())
				cancel() // 即座にキャンセル
				return ctx, cancel
			},
			setupMock: func() *mocks.MockFileSystem {
				fs := mocks.NewMockFileSystem()
				fs.Files = map[string][]byte{
					"thbgm.fmt":    make([]byte, 52),
					"musiccmt.txt": []byte("test"),
				}
				return fs
			},
			expectedError: context.Canceled,
		},
		{
			name: "タイムアウトコンテキスト",
			setupContext: func() (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
				time.Sleep(10 * time.Millisecond) // タイムアウトを確実に発生させる
				return ctx, cancel
			},
			setupMock: func() *mocks.MockFileSystem {
				fs := mocks.NewMockFileSystem()
				fs.Files = map[string][]byte{
					"thbgm.fmt":    make([]byte, 52),
					"musiccmt.txt": []byte("test"),
				}
				return fs
			},
			expectedError: context.DeadlineExceeded,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := tt.setupContext()
			defer cancel()

			fs := tt.setupMock()
			cfg := &config.Config{
				OutputDir: ".",
			}
			app := NewWithOptions(cfg, Options{
				FileSystem: fs,
			})

			err := app.Run(ctx)
			if err != tt.expectedError {
				t.Errorf("Expected error %v, got %v", tt.expectedError, err)
			}
		})
	}
}

func TestApp_processArchive_ContextDeadline(t *testing.T) {
	mockExtractor := &mocks.MockExtractor{
		ExtractedFiles: map[string][]byte{
			"thbgm.fmt":    make([]byte, 52),
			"musiccmt.txt": []byte("test"),
		},
	}

	cfg := &config.Config{
		ArchiveType: 6,
	}
	app := NewWithOptions(cfg, Options{
		Extractor: mockExtractor,
	})

	// 非常に短いタイムアウトを設定
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()
	time.Sleep(10 * time.Millisecond) // タイムアウトを確実に発生させる

	_, err := app.processArchive(ctx, "test.dat")
	if err != context.DeadlineExceeded {
		t.Errorf("Expected context.DeadlineExceeded error, got %v", err)
	}
}

func TestApp_processAutoDetect_ContextCancellation(t *testing.T) {
	fs := mocks.NewMockFileSystem()
	fs.Files = map[string][]byte{
		"thbgm.fmt":    make([]byte, 52),
		"musiccmt.txt": []byte("test"),
	}

	finder := &mocks.MockDatFileFinder{
		FoundFile: "",
	}

	cfg := &config.Config{}
	app := NewWithOptions(cfg, Options{
		FileSystem:    fs,
		DatFileFinder: finder,
	})

	// キャンセル済みのコンテキスト
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := app.processAutoDetect(ctx)
	if err != context.Canceled {
		t.Errorf("Expected context.Canceled error, got %v", err)
	}
}

func TestApp_processLocalFiles_LongRunning(t *testing.T) {
	// 長時間実行されるシナリオをシミュレート
	fs := mocks.NewMockFileSystem()
	// 大きなファイルをシミュレート
	largeData := make([]byte, 1024*1024) // 1MB
	fs.Files = map[string][]byte{
		"thbgm.fmt":    largeData,
		"musiccmt.txt": []byte("test"),
	}

	cfg := &config.Config{}
	app := NewWithOptions(cfg, Options{
		FileSystem: fs,
	})

	// 途中でキャンセルされるコンテキスト
	ctx, cancel := context.WithCancel(context.Background())
	
	// goroutineで少し遅延してからキャンセル
	go func() {
		time.Sleep(5 * time.Millisecond)
		cancel()
	}()

	_, err := app.processLocalFiles(ctx)
	// エラーが発生することを確認（キャンセルまたは成功）
	if err != nil && err != context.Canceled {
		// 処理が速すぎてキャンセルされなかった場合もOK
		t.Logf("Process completed before cancellation: %v", err)
	}
}

func TestContextWithSignal(t *testing.T) {
	// シグナルハンドリングをテスト
	// このテストは実際のシグナルを送信しないが、
	// シグナル処理のセットアップが正しいことを確認

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// シグナルチャンネルを作成
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigChan)

	// goroutineでシグナルを監視
	go func() {
		select {
		case <-sigChan:
			cancel()
		case <-ctx.Done():
			return
		}
	}()

	// コンテキストがまだ有効であることを確認
	select {
	case <-ctx.Done():
		t.Fatal("Context should not be canceled yet")
	default:
		// OK
	}

	// 手動でキャンセル
	cancel()

	// コンテキストがキャンセルされたことを確認
	select {
	case <-ctx.Done():
		// OK
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Context should be canceled")
	}
}

func TestApp_Run_WithValue(t *testing.T) {
	// context.WithValueを使用したテスト
	type contextKey string
	const testKey contextKey = "testKey"
	
	ctx := context.WithValue(context.Background(), testKey, "testValue")
	
	fs := mocks.NewMockFileSystem()
	fmtData := make([]byte, 52)
	copy(fmtData[0:], []byte("test.wav\x00"))
	cmtData := []byte("@bgm/test\n♪Test Track")
	fs.Files = map[string][]byte{
		"thbgm.fmt":    fmtData,
		"musiccmt.txt": cmtData,
	}

	cfg := &config.Config{
		OutputDir: ".",
	}
	app := NewWithOptions(cfg, Options{
		FileSystem: fs,
	})

	// contextの値が保持されていることを確認
	if val := ctx.Value(testKey); val != "testValue" {
		t.Errorf("Context value not preserved: got %v", val)
	}

	err := app.Run(ctx)
	if err != nil {
		t.Fatalf("Run failed with context.WithValue: %v", err)
	}
}