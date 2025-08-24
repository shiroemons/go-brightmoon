// Package parser はデータの解析を行います
package parser

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/shiroemons/go-brightmoon/internal/titles/fileutil"
	"github.com/shiroemons/go-brightmoon/internal/titles/models"
)

// THBGMParser はTHBGMフォーマットとmusiccmtファイルを解析します
type THBGMParser struct{}

// NewTHBGMParser は新しいTHBGMParserを作成します
func NewTHBGMParser() *THBGMParser {
	return &THBGMParser{}
}

// ParseTHFmt はthbgm.fmtファイルを解析します
func (p *THBGMParser) ParseTHFmt(data []byte) ([]*models.Record, error) {
	var records []*models.Record
	offset := 0

	for offset < len(data) {
		if offset+52 > len(data) {
			break
		}
		pcmFmt := data[offset : offset+52]
		fileBytes := pcmFmt[0:16]
		n := bytes.IndexByte(fileBytes, 0)
		file := string(fileBytes[:n])
		if file != "" {
			start := binary.LittleEndian.Uint32(pcmFmt[16:])
			intro := binary.LittleEndian.Uint32(pcmFmt[24:])
			length := binary.LittleEndian.Uint32(pcmFmt[28:])
			loop := length - intro
			record := &models.Record{
				FileName: file,
				Start:    toHex(start),
				Intro:    toHex(intro),
				Loop:     toHex(loop),
				Length:   toHex(length),
			}
			records = append(records, record)
		}
		offset += 52
	}

	return records, nil
}

// ParseMusicCmt はmusiccmt.txtファイルを解析します
func (p *THBGMParser) ParseMusicCmt(data string) ([]*models.Track, error) {
	// Shift-JISからUTF-8に変換
	text, err := fileutil.FromShiftJIS(data)
	if err != nil {
		return nil, fmt.Errorf("文字コード変換エラー: %w", err)
	}

	var tracks []*models.Track
	var fileName string
	var title string

	buf := bytes.NewBufferString(text)
	scanner := bufio.NewScanner(buf)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "@bgm/") {
			fileName = strings.Replace(line, "@bgm/", "", -1)
			if !strings.HasSuffix(fileName, ".wav") {
				fileName = fileName + ".wav"
			}
		}
		if strings.HasPrefix(line, "♪") {
			title = strings.Replace(line, "♪", "", -1)
			track := &models.Track{
				FileName: fileName,
				Title:    title,
			}
			tracks = append(tracks, track)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("スキャンエラー: %w", err)
	}

	return tracks, nil
}

// toHex はuint32を16進文字列に変換します
func toHex(i uint32) string {
	return fmt.Sprintf("%08X", i)
}
