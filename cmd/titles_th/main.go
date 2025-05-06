//
// GOOS=windows GOARCH=amd64 go build -o titles_th.exe main.go
//

package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

type Record struct {
	FileName string
	Start    string
	Intro    string
	Loop     string
	Length   string
}

type Track struct {
	FileName string
	Title    string
}

func exists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func toHex(i uint32) string {
	return fmt.Sprintf("%08X", i)
}

func transformEncoding(rawReader io.Reader, trans transform.Transformer) (string, error) {
	ret, err := io.ReadAll(transform.NewReader(rawReader, trans))
	if err == nil {
		return string(ret), nil
	} else {
		return "", err
	}
}

func FromShiftJIS(str string) (string, error) {
	return transformEncoding(strings.NewReader(str), japanese.ShiftJIS.NewDecoder())
}

func main() {
	var (
		err      error
		thfmt    []byte
		musiccmt string
	)

	if exists("thbgm.fmt") && exists("musiccmt.txt") {
		thfmt, err = os.ReadFile("thbgm.fmt")
		check(err)
		musiccmtBytes, err := os.ReadFile("musiccmt.txt")
		check(err)
		musiccmt = string(musiccmtBytes)
	} else if exists("thbgm_tr.fmt") && exists("musiccmt_tr.txt") {
		thfmt, err = os.ReadFile("thbgm_tr.fmt")
		check(err)
		musiccmtBytes, err := os.ReadFile("musiccmt_tr.txt")
		check(err)
		musiccmt = string(musiccmtBytes)
	} else {
		fmt.Println("thbgm.fmt, musiccmt.txt のファイルがありません。")
		os.Exit(1)
	}

	offset := 0
	var records []*Record
	for offset < len(thfmt) {
		if offset+52 > len(thfmt) {
			break
		}
		pcmFmt := thfmt[offset : offset+52]
		fileBytes := pcmFmt[0:16]
		n := bytes.IndexByte(fileBytes, 0)
		file := string(fileBytes[:n])
		if file != "" {
			start := binary.LittleEndian.Uint32(pcmFmt[16:])
			intro := binary.LittleEndian.Uint32(pcmFmt[24:])
			length := binary.LittleEndian.Uint32(pcmFmt[28:])
			loop := length - intro
			record := &Record{
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

	text, err := FromShiftJIS(musiccmt)
	check(err)

	var (
		fileName string
		title    string
	)

	var tracks []*Track

	buf := bytes.NewBufferString(text)
	scanner := bufio.NewScanner(buf)
	for scanner.Scan() {
		if strings.HasPrefix(scanner.Text(), "@bgm/") {
			fileName = strings.Replace(scanner.Text(), "@bgm/", "", -1)
			if !strings.HasSuffix(fileName, ".wav") {
				fileName = fileName + ".wav"
			}
		}
		if strings.HasPrefix(scanner.Text(), "♪") {
			title = strings.Replace(scanner.Text(), "♪", "", -1)
			track := &Track{
				FileName: fileName,
				Title:    title,
			}
			tracks = append(tracks, track)
		}
	}

	for _, t := range tracks {
		for _, r := range records {
			if t.FileName == r.FileName {
				fmt.Printf("%s,%s,%s,%s\n", r.Start, r.Intro, r.Loop, t.Title)
			}
		}
	}
	for i, r := range records {
		if len(tracks) <= i {
			if r.FileName == "th128_08.wav" {
				fmt.Printf("%s,%s,%s,%s\n", r.Start, r.Intro, r.Loop, "プレイヤーズスコア")
			} else {
				fmt.Printf("%s,%s,%s,%s\n", r.Start, r.Intro, r.Loop, r.FileName)
			}
		}
	}
}
