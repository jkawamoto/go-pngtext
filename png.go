// png.go
//
// Copyright (c) 2023 Junpei Kawamoto
//
// This software is released under the MIT License.
//
// http://opensource.org/licenses/mit-license.php

/*
Package pngtext provides function ParseTextualData that parses a PNG file and returns a list of textual data stored
in the file.

To parse textual data, simply call ParseTextualData:

	r, _ := os.Open("test.png")
	defer r.Close()

	res, _ := pngtext.ParseTextualData(r)

TextualDataList, which ParseTextualData returns as a result, is a list of TextualData but also implements Find function
that helps you to get text data associated with a keyword. This returns TextualData of which keyword is Description:

	res.Find("Description")
*/
package pngtext

import (
	"bufio"
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
	"strings"
)

const (
	pngHeader  = "\x89PNG\r\n\x1a\n"
	bufSize    = 8
	lengthSize = 4
	typeSize   = 4
	crcSize    = 4
)

var (
	ErrNotPngData                 = fmt.Errorf("not PNG data")
	ErrUnsupportedCompressionType = fmt.Errorf("unsupported compression type")
	ErrCRC                        = fmt.Errorf("CRC doesn't match")
)

// TextualData defines attributes one chunk could have.
// See also: https://www.w3.org/TR/2003/REC-PNG-20031110/#11textinfo
type TextualData struct {
	// Keyword of the textual data.
	Keyword string
	// Text string associated with the keyword
	Text string
	// LanguageTag indicates the human language used by the translated keyword and the text.
	// Only iTXt chunk has this attribute.
	LanguageTag string
	// TranslatedKeyword is a translation of the keyword into the language indicated by the language tag.
	// Only iTXt chunk has this attribute.
	TranslatedKeyword string
}

// TextualDataList is a list of *TextualData that provides Find and implements sort.Interface.
type TextualDataList []*TextualData

// Find sequentially searches an item of which keyword matches the given one. Returns nil if not matches any item.
func (list TextualDataList) Find(keyword string) *TextualData {
	for _, v := range list {
		if v.Keyword == keyword {
			return v
		}
	}
	return nil
}

// Len is the number of elements in the collection.
func (list TextualDataList) Len() int {
	return len(list)
}

// Less reports whether the element with index i
// must sort before the element with index j.
func (list TextualDataList) Less(i, j int) bool {
	return strings.Compare(list[i].Keyword, list[j].Keyword) < 0
}

// Swap swaps the elements with indexes i and j.
func (list TextualDataList) Swap(i, j int) {
	list[i], list[j] = list[j], list[i]
}

// ParseTextualData reads PNG data from the given reader and parses textual data.
func ParseTextualData(r io.Reader) (TextualDataList, error) {
	buf := make([]byte, bufSize)
	if _, err := io.ReadFull(r, buf[:len(pngHeader)]); err != nil {
		return nil, err
	}
	if string(buf[:len(pngHeader)]) != pngHeader {
		return nil, ErrNotPngData
	}

	var res []*TextualData
	for {
		if _, err := io.ReadFull(r, buf[:lengthSize]); err != nil {
			return nil, err
		}
		size := int64(binary.BigEndian.Uint32(buf[:lengthSize]))

		crc := crc32.NewIEEE()
		if _, err := io.ReadFull(io.TeeReader(r, crc), buf[:typeSize]); err != nil {
			return nil, err
		}
		chunkType := string(buf[:typeSize])

		data := bufio.NewReader(io.TeeReader(io.LimitReader(r, size), crc))
		switch chunkType {
		case "tEXt":
			v, err := parseTextData(data)
			if err != nil {
				return nil, err
			}
			res = append(res, v)

		case "zTXt":
			v, err := parseCompressedTextData(data)
			if err != nil {
				return nil, err
			}
			res = append(res, v)

		case "iTXt":
			v, err := parseInternationalTextData(data)
			if err != nil {
				return nil, err
			}
			res = append(res, v)

		default:
			_, err := io.Copy(io.Discard, data)
			if err != nil {
				return nil, err
			}
		}

		// check if CRC matches.
		if _, err := io.ReadFull(r, buf[:crcSize]); err != nil {
			return nil, err
		} else if !bytes.Equal(buf[:crcSize], crc.Sum(nil)) {
			return nil, ErrCRC
		}

		// check if last chunk is read.
		if chunkType == "IEND" {
			return res, nil
		}
	}
}

func trimTailingNull(s string) string {
	return s[:len(s)-1]
}

func parseTextData(r *bufio.Reader) (*TextualData, error) {
	keyword, err := r.ReadString(0)
	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, fmt.Errorf("failed to read keyword: %w", err)
	}

	value, err := r.ReadString(0)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed to read value: %w", err)
	}

	return &TextualData{
		Keyword: trimTailingNull(keyword),
		Text:    value,
	}, nil
}

func parseCompressedTextData(r *bufio.Reader) (*TextualData, error) {
	keyword, err := r.ReadString(0)
	if err != nil {
		return nil, fmt.Errorf("failed to read keyword: %w", err)
	}

	b, err := r.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("failed to read compression type: %w", err)
	} else if b != 0 {
		return nil, ErrUnsupportedCompressionType
	}

	zr, err := zlib.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("failed to decompress value: %w", err)
	}

	data, err := io.ReadAll(zr)
	if err != nil {
		return nil, fmt.Errorf("failed to read value: %w", err)
	}

	return &TextualData{
		Keyword: trimTailingNull(keyword),
		Text:    string(data),
	}, nil
}

func parseInternationalTextData(r *bufio.Reader) (*TextualData, error) {
	keyword, err := r.ReadString(0)
	if err != nil {
		return nil, fmt.Errorf("failed to read keyword: %w", err)
	}

	b, err := r.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("failed to read compression flag: %w", err)
	}
	compression := b == 1

	b, err = r.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("failed to read compression type: %w", err)
	} else if compression && b != 0 {
		return nil, ErrUnsupportedCompressionType
	}

	lang, err := r.ReadString(0)
	if err != nil {
		return nil, fmt.Errorf("failed to read language tag: %w", err)
	}

	translatedKeyword, err := r.ReadString(0)
	if err != nil {
		return nil, fmt.Errorf("failed to read translated keyword: %w", err)
	}

	var reader io.Reader = r
	if compression {
		reader, err = zlib.NewReader(r)
		if err != nil {
			return nil, fmt.Errorf("failed to decompress value: %w", err)
		}
	}
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read value: %w", err)
	}

	return &TextualData{
		Keyword:           trimTailingNull(keyword),
		Text:              string(data),
		LanguageTag:       trimTailingNull(lang),
		TranslatedKeyword: trimTailingNull(translatedKeyword),
	}, nil
}
