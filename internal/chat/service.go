package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"os"

	f "nordik-drive-api/internal/file"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
	"gorm.io/gorm"
)

type ChatService struct {
	DB     *gorm.DB
	APIKey string
}

func (cs *ChatService) Chat(question string, audioFile *multipart.FileHeader, filename string) (string, error) {
	// Fetch latest file version
	var file f.File
	if err := cs.DB.Where("filename = ?", filename).Order("version DESC").First(&file).Error; err != nil {
		return "", fmt.Errorf("file not found")
	}

	var fileData []f.FileData
	if err := cs.DB.Where("file_id = ? AND version = ?", file.ID, file.Version).Find(&fileData).Error; err != nil {
		return "", fmt.Errorf("file data not found")
	}

	var allRows []json.RawMessage
	for _, row := range fileData {
		allRows = append(allRows, json.RawMessage(row.RowData))
	}
	fileDataJSON, err := json.Marshal(allRows)
	if err != nil {
		return "", fmt.Errorf("failed to marshal file data: %w", err)
	}

	ctx := context.Background()

	// Load GEMINI_KEY from env if not provided
	apiKey := cs.APIKey
	if apiKey == "" {
		apiKey = os.Getenv("GEMINI_KEY")
	}
	if apiKey == "" {
		return "", fmt.Errorf("missing GEMINI_KEY")
	}

	// Create Gemini client (public API)
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return "", fmt.Errorf("failed to create Gemini client: %w", err)
	}
	defer client.Close()

	// Pick correct model
	var model *genai.GenerativeModel
	if audioFile != nil {
		model = client.GenerativeModel("gemini-2.5-flash") // multimodal flash
	} else {
		model = client.GenerativeModel("gemini-2.5-flash") // text-only flash
	}

	// Compose prompt
	prompt := question + "\n\nFile name: " + filename +
		"\n\nAnswer the question based on the provided data, no json format response please: " + string(fileDataJSON)

	var response string

	if audioFile != nil {
		// Read audio
		fh, err := audioFile.Open()
		if err != nil {
			return "", fmt.Errorf("failed to open audio file: %w", err)
		}
		defer fh.Close()

		audioBytes, err := io.ReadAll(fh)
		if err != nil {
			return "", fmt.Errorf("failed to read audio file: %w", err)
		}

		audioMimeType := audioFile.Header.Get("Content-Type")
		if audioMimeType == "application/octet-stream" {
			audioMimeType = "audio/webm"
		}

		// Send multimodal request
		resp, err := model.GenerateContent(ctx,
			genai.Text(prompt),
			genai.Blob{MIMEType: audioMimeType, Data: audioBytes},
		)
		if err != nil {
			return "", fmt.Errorf("generation error: %w", err)
		}

		for _, cand := range resp.Candidates {
			if cand.Content != nil {
				for _, part := range cand.Content.Parts {
					if text, ok := part.(genai.Text); ok {
						response = string(text)
						break
					}
				}
			}
		}

	} else {
		// Text-only request
		resp, err := model.GenerateContent(ctx, genai.Text(prompt))
		if err != nil {
			return "", fmt.Errorf("generation error: %w", err)
		}

		for _, cand := range resp.Candidates {
			if cand.Content != nil {
				for _, part := range cand.Content.Parts {
					if text, ok := part.(genai.Text); ok {
						response = string(text)
						break
					}
				}
			}
		}
	}

	if response == "" {
		return "", fmt.Errorf("no response from Gemini")
	}

	return response, nil
}
