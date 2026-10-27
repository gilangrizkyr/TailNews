package platform

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/image/draw"
	_ "golang.org/x/image/webp"
)

var (
	ErrInvalidImageType = errors.New("unsupported image format; only JPEG, PNG, and WebP are allowed")
	ErrImageTooLarge    = errors.New("image size exceeds maximum allowed limit")
)

type ProcessedMedia struct {
	OriginalFilename string
	StoredFilename   string
	FilePath         string
	MimeType         string
	SizeBytes        uint64
	Width            uint32
	Height           uint32
	Variants         map[string]string // e.g., "thumbnail": "/media/...", "medium": "...", "large": "..."
}

type StoragePlatform interface {
	SaveImage(r io.Reader, originalFilename string, maxBytes int64) (*ProcessedMedia, error)
	DeleteImage(filePath string, variants map[string]string) error
}

type LocalStorage struct {
	baseDir string
}

func NewLocalStorage(baseDir string) (*LocalStorage, error) {
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to initialize media directory: %w", err)
	}
	return &LocalStorage{baseDir: baseDir}, nil
}

func (s *LocalStorage) SaveImage(r io.Reader, originalFilename string, maxBytes int64) (*ProcessedMedia, error) {
	// Read header bytes to detect magic MIME type
	header := make([]byte, 512)
	n, err := r.Read(header)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed to read file header: %w", err)
	}

	mimeType := http.DetectContentType(header[:n])
	isWebP := bytes.HasPrefix(header[:n], []byte("RIFF")) && len(header) >= 12 && string(header[8:12]) == "WEBP"

	if !strings.HasPrefix(mimeType, "image/jpeg") &&
		!strings.HasPrefix(mimeType, "image/png") &&
		!isWebP {
		return nil, ErrInvalidImageType
	}

	// Read full file into memory buffer up to maxBytes limit
	var fullBuffer bytes.Buffer
	fullBuffer.Write(header[:n])

	limitedReader := io.LimitReader(r, maxBytes-int64(n)+1)
	copied, err := io.Copy(&fullBuffer, limitedReader)
	if err != nil {
		return nil, fmt.Errorf("error reading file payload: %w", err)
	}
	if int64(n)+copied > maxBytes {
		return nil, ErrImageTooLarge
	}

	// Decode image (stripping EXIF metadata automatically)
	img, _, err := image.Decode(bytes.NewReader(fullBuffer.Bytes()))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image data: %w", err)
	}

	bounds := img.Bounds()
	origWidth := uint32(bounds.Dx())
	origHeight := uint32(bounds.Dy())

	// Generate safe unique filename
	randomBytes := make([]byte, 12)
	_, _ = rand.Read(randomBytes)
	baseName := fmt.Sprintf("%d_%s", time.Now().Unix(), hex.EncodeToString(randomBytes))

	// Save original image as high quality JPEG
	origFilename := fmt.Sprintf("%s_orig.jpg", baseName)
	origFilePath := filepath.Join(s.baseDir, origFilename)
	origFile, err := os.Create(origFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create original file: %w", err)
	}
	defer origFile.Close()

	if err := jpeg.Encode(origFile, img, &jpeg.Options{Quality: 90}); err != nil {
		return nil, fmt.Errorf("failed to encode original image: %w", err)
	}

	fi, _ := origFile.Stat()
	totalBytes := uint64(fi.Size())

	// Generate resized variants: large (1200px), medium (800px), thumbnail (400px)
	variants := make(map[string]string)
	variantSizes := map[string]int{
		"large":     1200,
		"medium":    800,
		"thumbnail": 400,
	}

	for name, targetWidth := range variantSizes {
		resized := resizeImage(img, targetWidth)
		varFilename := fmt.Sprintf("%s_%s.jpg", baseName, name)
		varPath := filepath.Join(s.baseDir, varFilename)

		varFile, err := os.Create(varPath)
		if err != nil {
			continue
		}
		_ = jpeg.Encode(varFile, resized, &jpeg.Options{Quality: 85})
		varFile.Close()

		variants[name] = "/media/" + varFilename
	}

	return &ProcessedMedia{
		OriginalFilename: originalFilename,
		StoredFilename:   origFilename,
		FilePath:         "/media/" + origFilename,
		MimeType:         "image/jpeg",
		SizeBytes:        totalBytes,
		Width:            origWidth,
		Height:           origHeight,
		Variants:         variants,
	}, nil
}

func (s *LocalStorage) DeleteImage(filePath string, variants map[string]string) error {
	base := filepath.Base(filePath)
	_ = os.Remove(filepath.Join(s.baseDir, base))

	for _, vPath := range variants {
		vBase := filepath.Base(vPath)
		_ = os.Remove(filepath.Join(s.baseDir, vBase))
	}
	return nil
}

func resizeImage(src image.Image, targetWidth int) image.Image {
	bounds := src.Bounds()
	srcWidth := bounds.Dx()
	srcHeight := bounds.Dy()

	if srcWidth <= targetWidth {
		return src
	}

	targetHeight := (srcHeight * targetWidth) / srcWidth
	dst := image.NewRGBA(image.Rect(0, 0, targetWidth, targetHeight))
	draw.BiLinear.Scale(dst, dst.Bounds(), src, bounds, draw.Over, nil)
	return dst
}

// Compile-time check
var _ StoragePlatform = (*LocalStorage)(nil)
