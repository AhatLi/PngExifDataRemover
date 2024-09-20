package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

const pngSignature = "\x89PNG\r\n\x1a\n"

type chunk struct {
	Length uint32
	Type   [4]byte
	Data   []byte
	CRC    uint32
}

func main() {
	// 출력 디렉토리 생성
	err := os.MkdirAll(GetConfig().InputDir, os.ModePerm)
	if err != nil {
		panic(err)
	}

	err = filepath.Walk(GetConfig().InputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.ToLower(filepath.Ext(path)) == ".png" {
			err := processPNGFile(path, filepath.Join(GetConfig().OutputDir, info.Name()))
			if err != nil {
				fmt.Printf("Error processing %s: %v\n", path, err)
			}
		}
		return nil
	})

	if err != nil {
		panic(err)
	}
	fmt.Println("All PNG files have been processed.")
}

func processPNGFile(inputPath string, outputPath string) error {

	// PNG 파일 읽기
	pngBytes, err := ioutil.ReadFile(inputPath)
	if err != nil {
		panic(err)
	}

	// PNG 시그니처 확인
	if string(pngBytes[:8]) != pngSignature {
		panic("Invalid PNG file")
	}

	// 청크 파싱 및 수정
	var modifiedChunks []chunk
	r := bytes.NewReader(pngBytes[8:])
	for {
		c, err := readChunk(r)
		if err != nil {
			break
		}

		if string(c.Type[:]) == "iTXt" || string(c.Type[:]) == "tEXt" {
			modifiedData := modifyTextChunk(c.Data)
			c.Length = uint32(len(modifiedData))
			c.Data = modifiedData
			c.CRC = calculateCRC(c.Type[:], c.Data)
		}

		modifiedChunks = append(modifiedChunks, c)
	}

	// 수정된 PNG 파일 생성
	var buf bytes.Buffer
	buf.WriteString(pngSignature)
	for _, c := range modifiedChunks {
		writeChunk(&buf, c)
	}

	// 새 파일로 저장
	err = ioutil.WriteFile(outputPath, buf.Bytes(), 0644)
	if err != nil {
		panic(err)
	}

	return nil
}

func readChunk(r *bytes.Reader) (chunk, error) {
	var c chunk
	if err := binary.Read(r, binary.BigEndian, &c.Length); err != nil {
		return c, err
	}
	if err := binary.Read(r, binary.BigEndian, &c.Type); err != nil {
		return c, err
	}
	c.Data = make([]byte, c.Length)
	if _, err := r.Read(c.Data); err != nil {
		return c, err
	}
	if err := binary.Read(r, binary.BigEndian, &c.CRC); err != nil {
		return c, err
	}
	return c, nil
}

func writeChunk(w *bytes.Buffer, c chunk) {
	binary.Write(w, binary.BigEndian, c.Length)
	w.Write(c.Type[:])
	w.Write(c.Data)
	binary.Write(w, binary.BigEndian, c.CRC)
}

func modifyTextChunk(data []byte) []byte {
	parts := bytes.SplitN(data, []byte{0}, 2)
	if len(parts) != 2 {
		return data
	}
	key, value := parts[0], parts[1]

	for _, data := range GetConfig().RemoveString {
		value = bytes.Replace(value, []byte(data), nil, -1)
	}

	return append(append(key, 0), value...)
}

func calculateCRC(data ...[]byte) uint32 {
	crc := uint32(0xffffffff)
	for _, d := range data {
		crc = updateCRC(crc, d)
	}
	return crc ^ 0xffffffff
}

func updateCRC(crc uint32, data []byte) uint32 {
	for _, b := range data {
		crc = crcTable[(crc^uint32(b))&0xff] ^ (crc >> 8)
	}
	return crc
}

var crcTable = make([]uint32, 256)

func init() {
	for n := 0; n < 256; n++ {
		c := uint32(n)
		for k := 0; k < 8; k++ {
			if c&1 != 0 {
				c = 0xedb88320 ^ (c >> 1)
			} else {
				c = c >> 1
			}
		}
		crcTable[n] = c
	}
}
