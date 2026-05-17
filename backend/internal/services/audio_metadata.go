package services

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/dhowden/tag"
)

type AudioMetadata struct {
	Title       string
	Artist      string
	Album       string
	Year        string
	ReleaseDate string
	Genre       string
	CoverURL    string
	DurationSec int
}

var filenameArtistTitleRe = regexp.MustCompile(`^(.+?)\s*[-–—－]\s*(.+)$`)
var trackNumberPrefixRe = regexp.MustCompile(`^\d{1,3}[\.\s、]\s*`)

func ParseFilenameMetadata(filename string) (artist, title string) {
	name := strings.TrimSuffix(filename, filepath.Ext(filename))
	if m := filenameArtistTitleRe.FindStringSubmatch(name); m != nil {
		return strings.TrimSpace(m[1]), strings.TrimSpace(m[2])
	}
	return "", strings.TrimSpace(trackNumberPrefixRe.ReplaceAllString(name, ""))
}

// ReadAudioMetadata 从音频文件读取嵌入的元数据（ID3 标签等）。
// originalName 是用户上传的原始文件名，当 tag 读取失败时用于从文件名解析 artist/title。
func ReadAudioMetadata(path string, originalName string) AudioMetadata {
	ext := filepath.Ext(path)
	dur := estimateDuration(path, ext)

	file, err := os.Open(path)
	if err != nil {
		log.Printf("music metadata: cannot open file %s: %v", path, err)
		return fallbackMetadataFromFilename(originalName, dur.DurationSec)
	}
	defer file.Close()

	metadata, err := tag.ReadFrom(file)
	if err != nil {
		log.Printf("music metadata: tag read failed for %s: %v", path, err)
		result := fallbackMetadataFromFilename(originalName, dur.DurationSec)
		result.CoverURL = ExtractEmbeddedCover(path)
		return result
	}

	result := AudioMetadata{
		Title:       strings.TrimSpace(metadata.Title()),
		Artist:      strings.TrimSpace(firstNonEmptyString(metadata.Artist(), metadata.AlbumArtist(), metadata.Composer())),
		Album:       strings.TrimSpace(metadata.Album()),
		Genre:       strings.TrimSpace(metadata.Genre()),
		DurationSec: dur.DurationSec,
	}
	if year := metadata.Year(); year > 0 {
		result.Year = strconv.Itoa(year)
	}

	if pic := metadata.Picture(); pic != nil && len(pic.Data) > 0 {
		result.CoverURL = writeCoverFile(pic.Data, pic.Ext)
	}
	if result.CoverURL == "" {
		result.CoverURL = ExtractEmbeddedCover(path)
	}

	// 如果 tag 读取成功但 title/artist 为空（有些文件 tag 存在但字段为空），
	// 使用文件名解析作为补充
	if result.Title == "" || result.Artist == "" {
		fnArtist, fnTitle := ParseFilenameMetadata(originalName)
		if result.Title == "" {
			result.Title = fnTitle
		}
		if result.Artist == "" {
			result.Artist = fnArtist
		}
	}

	return result
}

// ExtractEmbeddedCover extracts an attached picture stream that tag readers may miss.
// Some MP3 files store album art as an mjpeg/png video stream; ffmpeg handles that
// format reliably when it is available on the host.
func ExtractEmbeddedCover(path string) string {
	if path == "" {
		return ""
	}
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		return ""
	}

	coverPath := filepath.Join("uploads", "music", "covers", fmt.Sprintf("%d.jpg", time.Now().UnixNano()))
	if err := os.MkdirAll(filepath.Dir(coverPath), 0755); err != nil {
		log.Printf("music metadata: cannot create covers directory: %v", err)
		return ""
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "ffmpeg", "-y", "-v", "error", "-i", path, "-map", "0:v:0", "-frames:v", "1", coverPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		_ = os.Remove(coverPath)
		if strings.TrimSpace(string(output)) != "" {
			log.Printf("music metadata: ffmpeg cover extraction failed for %s: %v: %s", path, err, strings.TrimSpace(string(output)))
		}
		return ""
	}
	if info, err := os.Stat(coverPath); err != nil || info.Size() == 0 {
		_ = os.Remove(coverPath)
		return ""
	}
	return "/" + filepath.ToSlash(coverPath)
}

func writeCoverFile(data []byte, ext string) string {
	ext = strings.TrimPrefix(strings.TrimSpace(ext), ".")
	if ext == "" {
		ext = "jpg"
	}
	coverDir := filepath.Join("uploads", "music", "covers")
	if err := os.MkdirAll(coverDir, 0755); err != nil {
		log.Printf("music metadata: cannot create covers directory: %v", err)
		return ""
	}
	coverPath := filepath.Join(coverDir, fmt.Sprintf("%d.%s", time.Now().UnixNano(), ext))
	if err := os.WriteFile(coverPath, data, 0644); err != nil {
		log.Printf("music metadata: cannot write cover image: %v", err)
		return ""
	}
	return "/" + filepath.ToSlash(coverPath)
}

// fallbackMetadataFromFilename 当音频标签读取失败时，从原始文件名解析元数据
func fallbackMetadataFromFilename(originalName string, durationSec int) AudioMetadata {
	artist, title := ParseFilenameMetadata(originalName)
	return AudioMetadata{
		Title:       title,
		Artist:      artist,
		DurationSec: durationSec,
	}
}

func estimateDuration(path string, ext string) AudioMetadata {
	switch strings.ToLower(ext) {
	case ".mp3":
		return AudioMetadata{DurationSec: estimateMP3Duration(path)}
	case ".flac":
		return readContainerDuration(path)
	default:
		return readContainerDuration(path)
	}
}

func estimateMP3Duration(path string) int {
	f, err := os.Open(path)
	if err != nil {
		return 0
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return 0
	}
	fileSize := info.Size()

	header := make([]byte, 10)
	if _, err := f.Read(header); err != nil {
		return 0
	}

	dataStart := int64(0)
	if header[0] == 0x49 && header[1] == 0x44 && header[2] == 0x33 {
		if len(header) >= 10 {
			id3Size := ((int64(header[6]) & 0x7f) << 21) |
				((int64(header[7]) & 0x7f) << 14) |
				((int64(header[8]) & 0x7f) << 7) |
				(int64(header[9]) & 0x7f)
			dataStart = id3Size + 10
		}
	}

	if dataStart > 0 {
		if _, err := f.Seek(dataStart, 0); err != nil {
			return 0
		}
	}

	buf := make([]byte, 4096)
	tail := make([]byte, 0, 3)
	for {
		n, err := f.Read(buf)
		if n > 0 {
			data := append(tail, buf[:n]...)
			for i := 0; i <= len(data)-4; i++ {
				if data[i] == 0xFF && data[i+1]&0xE0 == 0xE0 {
					frameHeader := data[i : i+4]
					bitrate, sampleRate, samplesPerFrame, ok := parseMPEGFrame(frameHeader)
					if !ok {
						continue
					}
					frameSize := mpegFrameSize(frameHeader, bitrate, sampleRate)
					if frameSize <= 0 {
						continue
					}
					totalFrames := (fileSize - dataStart) / int64(frameSize)
					if totalFrames <= 0 {
						return 0
					}
					return int(int64(samplesPerFrame) * totalFrames / int64(sampleRate))
				}
			}
			tailLen := 3
			if len(data) < tailLen {
				tailLen = len(data)
			}
			tail = append(tail[:0], data[len(data)-tailLen:]...)
		}
		if err != nil {
			return 0
		}
		if n == 0 {
			return 0
		}
	}
}

func parseMPEGFrame(header []byte) (bitrate, sampleRate, samplesPerFrame int, ok bool) {
	if len(header) < 4 || header[0] != 0xFF || header[1]&0xE0 != 0xE0 {
		return 0, 0, 0, false
	}

	versionIndex := int((header[1] >> 3) & 0x03)
	layerIndex := int((header[1] >> 1) & 0x03)
	bitrateIndex := int((header[2] >> 4) & 0x0F)
	sampleRateIndex := int((header[2] >> 2) & 0x03)

	if versionIndex == 1 || layerIndex == 0 || bitrateIndex == 0 || bitrateIndex == 15 || sampleRateIndex == 3 {
		return 0, 0, 0, false
	}

	bitrateTable := mpegBitrateTable(versionIndex, layerIndex)
	sampleRateTable := mpegSampleRateTable(versionIndex)
	samplesTable := mpegSamplesTable(versionIndex, layerIndex)

	if bitrateIndex >= len(bitrateTable) || sampleRateIndex >= len(sampleRateTable) {
		return 0, 0, 0, false
	}

	return bitrateTable[bitrateIndex], sampleRateTable[sampleRateIndex], samplesTable, true
}

func mpegFrameSize(header []byte, bitrate, sampleRate int) int {
	if len(header) < 4 {
		return 0
	}
	versionIndex := int((header[1] >> 3) & 0x03)
	layerIndex := int((header[1] >> 1) & 0x03)
	padding := int((header[2] >> 1) & 0x01)

	if layerIndex == 3 {
		return 144*bitrate*1000/sampleRate + padding
	}
	if versionIndex == 3 {
		return 144*bitrate*1000/sampleRate + padding
	}
	return 72*bitrate*1000/sampleRate + padding
}

func mpegBitrateTable(version, layer int) []int {
	switch {
	case version == 3 && layer == 3:
		return []int{0, 32, 64, 96, 128, 160, 192, 224, 256, 288, 320, 352, 384, 416, 448, 0}
	case version == 3 && layer == 2:
		return []int{0, 32, 48, 56, 64, 80, 96, 112, 128, 160, 192, 224, 256, 320, 384, 0}
	case version == 3 && layer == 1:
		return []int{0, 32, 64, 96, 128, 160, 192, 224, 256, 288, 320, 352, 384, 416, 448, 0}
	case version != 3 && layer == 3:
		return []int{0, 8, 16, 24, 32, 40, 48, 56, 64, 80, 96, 112, 128, 144, 160, 0}
	case version != 3 && layer == 2:
		return []int{0, 8, 16, 24, 32, 40, 48, 56, 64, 80, 96, 112, 128, 144, 160, 0}
	default:
		return []int{0, 32, 48, 56, 64, 80, 96, 112, 128, 144, 160, 176, 192, 224, 256, 0}
	}
}

func mpegSampleRateTable(version int) []int {
	if version == 3 {
		return []int{44100, 48000, 32000, 0}
	}
	return []int{22050, 24000, 16000, 0}
}

func mpegSamplesTable(version, layer int) int {
	if layer == 3 {
		return 1152
	}
	return 384
}

func readContainerDuration(path string) AudioMetadata {
	f, err := os.Open(path)
	if err != nil {
		return AudioMetadata{}
	}
	defer f.Close()

	magic := make([]byte, 4)
	if _, err := io.ReadFull(f, magic); err != nil {
		return AudioMetadata{}
	}
	if string(magic) == "fLaC" {
		return readFLACDurationFromReader(f)
	}
	return AudioMetadata{}
}

func readFLACDurationFromReader(r io.Reader) AudioMetadata {
	headerBuf := make([]byte, 4)
	for {
		if _, err := io.ReadFull(r, headerBuf); err != nil {
			return AudioMetadata{}
		}
		header := headerBuf[0]
		blockType := header & 0x7f
		length := int(headerBuf[1])<<16 | int(headerBuf[2])<<8 | int(headerBuf[3])
		if blockType == 0 && length >= 18 {
			block := make([]byte, length)
			if _, err := io.ReadFull(r, block); err != nil {
				return AudioMetadata{}
			}
			sampleRate := int(block[10])<<12 | int(block[11])<<4 | int(block[12]>>4)
			totalSamples := uint64(block[13]&0x0f)<<32 | uint64(block[14])<<24 | uint64(block[15])<<16 | uint64(block[16])<<8 | uint64(block[17])
			if sampleRate > 0 && totalSamples > 0 {
				return AudioMetadata{DurationSec: int(totalSamples / uint64(sampleRate))}
			}
		} else {
			if _, err := io.CopyN(io.Discard, r, int64(length)); err != nil {
				return AudioMetadata{}
			}
		}
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
