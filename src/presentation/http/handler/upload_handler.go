package handler

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	miniostorage "github.com/enterprise/trade-license/src/infrastructure/storage/minio"
)

type UploadHandler struct {
	storage *miniostorage.Storage
}

func NewUploadHandler(storage *miniostorage.Storage) *UploadHandler {
	return &UploadHandler{storage: storage}
}

var allowedTypes = map[string]string{
	"application/pdf": ".pdf",
	"image/jpeg":      ".jpg",
	"image/png":       ".png",
	"image/webp":      ".webp",
}

// UploadDocument handles POST /api/customer/upload.
// Accepts a multipart form with a single "file" field.
// Returns {"key": "...", "name": "...", "content_type": "..."}.
func (h *UploadHandler) UploadDocument(c *fiber.Ctx) error {
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "file field is required"})
	}

	if file.Size > 20<<20 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "file exceeds 20 MB limit"})
	}

	contentType := file.Header.Get("Content-Type")
	if _, ok := allowedTypes[contentType]; !ok {
		ext := strings.ToLower(filepath.Ext(file.Filename))
		switch ext {
		case ".pdf":
			contentType = "application/pdf"
		case ".jpg", ".jpeg":
			contentType = "image/jpeg"
		case ".png":
			contentType = "image/png"
		case ".webp":
			contentType = "image/webp"
		default:
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "unsupported file type; allowed: pdf, jpg, png, webp",
			})
		}
	}

	src, err := file.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not read upload"})
	}
	defer src.Close()

	buf, err := io.ReadAll(src)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not read upload"})
	}

	ext := allowedTypes[contentType]
	objectKey := fmt.Sprintf("documents/%s%s", uuid.New().String(), ext)

	if err := h.storage.Upload(c.Context(), objectKey, bytes.NewReader(buf), int64(len(buf)), contentType); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "upload failed: " + err.Error()})
	}

	return c.JSON(fiber.Map{
		"key":          objectKey,
		"name":         file.Filename,
		"content_type": contentType,
	})
}

// StreamDocument handles GET /api/documents/view?key=<object_key>.
// Fetches the object from MinIO internally and streams it directly to the browser.
// This avoids presigned URL complexity and keeps MinIO off the public network.
func (h *UploadHandler) StreamDocument(c *fiber.Ctx) error {
	key := c.Query("key")
	if key == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "key query param is required"})
	}

	// Seed/demo data stores full http(s) URLs — proxy-fetch them so the
	// browser never needs direct access.
	if strings.HasPrefix(key, "http://") || strings.HasPrefix(key, "https://") {
		resp, err := http.Get(key) //nolint:noctx
		if err != nil || resp.StatusCode != http.StatusOK {
			return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": "could not fetch external document"})
		}
		defer resp.Body.Close()
		ct := resp.Header.Get("Content-Type")
		if ct == "" {
			ct = "application/octet-stream"
		}
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": "could not read external document"})
		}
		c.Set("Content-Type", ct)
		c.Set("Content-Disposition", "inline")
		return c.Send(data)
	}

	obj, err := h.storage.GetObject(c.Context(), key)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "document not found"})
	}
	defer obj.Close()

	stat, err := obj.Stat()
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "document not found"})
	}

	data, err := io.ReadAll(obj)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not read document"})
	}

	c.Set("Content-Type", stat.ContentType)
	c.Set("Content-Disposition", "inline")
	return c.Send(data)
}
