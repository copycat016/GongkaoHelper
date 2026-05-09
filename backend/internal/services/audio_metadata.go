package services

import (
	"encoding/binary"
	"fmt"
	"os"
	"strconv"

	"github.com/dhowden/tag"
)

type AudioMetadata struct {
	Title       string
	Artist      string
	Album       string
	Year        string
	Genre       string
	DurationSec int
}

func ReadAudioMetadata(path string) AudioMetadata {
	file, err := os.Open(path)
	if err != nil {
		return AudioMetadata{}
	}
	defer file.Close()

	metadata, err := tag.ReadFrom(file)
	if err != nil {
		return readContainerDuration(path)
	}

	result := AudioMetadata{
		Title:  metadata.Title(),
		Artist: firstNonEmptyString(metadata.Artist(), metadata.AlbumArtist(), metadata.Composer()),
		Album:  metadata.Album(),
		Genre:  metadata.Genre(),
	}
	if year := metadata.Year(); year > 0 {
		result.Year = strconv.Itoa(year)
	}
	if result.DurationSec == 0 {
		result.DurationSec = readContainerDuration(path).DurationSec
	}
	return result
}

func readContainerDuration(path string) AudioMetadata {
	data, err := os.ReadFile(path)
	if err != nil {
		return AudioMetadata{}
	}
	if len(data) >= 4 && string(data[:4]) == "fLaC" {
		return readFLACDuration(data)
	}
	return AudioMetadata{}
}

func readFLACDuration(data []byte) AudioMetadata {
	offset := 4
	for offset+4 <= len(data) {
		header := data[offset]
		blockType := header & 0x7f
		length := int(data[offset+1])<<16 | int(data[offset+2])<<8 | int(data[offset+3])
		offset += 4
		if offset+length > len(data) {
			break
		}
		block := data[offset : offset+length]
		if blockType == 0 && len(block) >= 18 {
			sampleRate := int(block[10])<<12 | int(block[11])<<4 | int(block[12]>>4)
			totalSamples := uint64(block[13]&0x0f)<<32 | uint64(block[14])<<24 | uint64(block[15])<<16 | uint64(block[16])<<8 | uint64(block[17])
			if sampleRate > 0 && totalSamples > 0 {
				return AudioMetadata{DurationSec: int(totalSamples / uint64(sampleRate))}
			}
		}
		offset += length
		if header&0x80 != 0 {
			break
		}
	}
	return AudioMetadata{}
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func durationFromRawLength(raw map[string]interface{}) int {
	value, ok := raw["Length"]
	if !ok {
		return 0
	}
	switch typed := value.(type) {
	case string:
		ms, err := strconv.Atoi(typed)
		if err == nil && ms > 0 {
			return ms / 1000
		}
	case []byte:
		if len(typed) >= 4 {
			return int(binary.BigEndian.Uint32(typed)) / 1000
		}
	default:
		_, _ = fmt.Sprint(typed), ok
	}
	return 0
}
