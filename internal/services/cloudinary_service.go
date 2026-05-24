package services

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"disaster_alert_backend/config"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

type CloudinaryService struct {
	client *cloudinary.Cloudinary
	folder string
}

type CloudinaryUploadResult struct {
	MediaURL      string `json:"mediaUrl"`
	MediaType     string `json:"mediaType"`
	MediaPublicID string `json:"mediaPublicId"`
}

func NewCloudinaryService(cfg *config.Config) (*CloudinaryService, error) {
	if cfg.CloudinaryCloudName == "" || cfg.CloudinaryAPIKey == "" || cfg.CloudinaryAPISecret == "" {
		return nil, fmt.Errorf("cloudinary credentials are missing")
	}

	cld, err := cloudinary.NewFromParams(
		cfg.CloudinaryCloudName,
		cfg.CloudinaryAPIKey,
		cfg.CloudinaryAPISecret,
	)
	if err != nil {
		return nil, err
	}

	folder := cfg.CloudinaryUploadFolder
	if folder == "" {
		folder = "disaster_alert/reports"
	}

	return &CloudinaryService{
		client: cld,
		folder: folder,
	}, nil
}

func (s *CloudinaryService) UploadReportMedia(
	ctx context.Context,
	file multipart.File,
	header *multipart.FileHeader,
) (*CloudinaryUploadResult, error) {
	if file == nil || header == nil {
		return nil, fmt.Errorf("media file is required")
	}

	if header.Size > 50*1024*1024 {
		return nil, fmt.Errorf("media file is too large; maximum size is 50MB")
	}

	mediaType, err := detectMediaType(header.Filename)
	if err != nil {
		return nil, err
	}

	publicID := fmt.Sprintf(
		"report_%d_%s",
		time.Now().UnixNano(),
		sanitizeFilename(header.Filename),
	)

	uploadResult, err := s.client.Upload.Upload(
		ctx,
		file,
		uploader.UploadParams{
			Folder:       s.folder,
			PublicID:     publicID,
			ResourceType: "auto",
		},
	)
	if err != nil {
		return nil, err
	}

	return &CloudinaryUploadResult{
		MediaURL:      uploadResult.SecureURL,
		MediaType:     mediaType,
		MediaPublicID: uploadResult.PublicID,
	}, nil
}

func detectMediaType(filename string) (string, error) {
	extension := strings.ToLower(filepath.Ext(filename))

	switch extension {
	case ".jpg", ".jpeg", ".png", ".webp":
		return "image", nil
	case ".mp4", ".mov", ".avi", ".mkv", ".webm":
		return "video", nil
	default:
		return "", fmt.Errorf("unsupported media type: %s", extension)
	}
}

func sanitizeFilename(filename string) string {
	name := strings.TrimSuffix(filename, filepath.Ext(filename))
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "-", "_")

	var builder strings.Builder

	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			builder.WriteRune(r)
		}
	}

	clean := builder.String()
	if clean == "" {
		return "media"
	}

	return clean
}