// Package models はtitles_thコマンドで使用するデータモデルを定義します
package models

// Record は音楽ファイルの再生情報を表します
type Record struct {
	FileName string
	Start    string // 開始位置（16進数）
	Intro    string // イントロ部の長さ（16進数）
	Loop     string // ループ部の長さ（16進数）
	Length   string // 全体の長さ（16進数）
}

// Track は音楽トラック情報を表します
type Track struct {
	FileName string
	Title    string
}

// AdditionalInfo は補足情報を保持します
type AdditionalInfo struct {
	HasAdditionalInfo bool
	TitleInfo         string
	DisplayTitle      string
	IsTrialVersion    bool
	Error             error
}

// ExtractedData はアーカイブから抽出したデータを表します
type ExtractedData struct {
	THFmt     []byte
	MusicCmt  string
	InputFile string
}
