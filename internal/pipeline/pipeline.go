package pipeline

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"parce/internal/llm"
	"parce/internal/model"
	"parce/internal/ocr"
	"parce/internal/pdfimage"
)

func IsPDF(data []byte) bool {
	return len(data) >= 4 && string(data[:4]) == "%PDF"
}

func ExtractViaVision(imageData []byte) (*model.Transaction, error) {
	apiKey := os.Getenv("GROQ_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GROQ_API_KEY is not set")
	}

	mimeType := http.DetectContentType(imageData)
	slog.Info("vision extraction", "size_bytes", len(imageData), "mime", mimeType)

	transaction, _, err := llm.NewGroqClient(apiKey).ExtractTransactionFromImage(imageData, mimeType)
	if err != nil {
		return nil, err
	}
	return transaction, nil
}


func ExtractViaPDF(pdfPath string) (*model.Transaction, error) {
	imageData, err := pdfimage.RenderFirstPage(pdfPath)
	if err != nil {
		return nil, fmt.Errorf("render pdf: %w", err)
	}
	return ExtractViaVision(imageData)
}


func ExtractViaOCR(imagePath string) (*model.Transaction, error) {
	rawText, err := ocr.ExtractText(imagePath)
	if err != nil {
		return nil, err
	}
	slog.Info("ocr done", "text_len", len(rawText))

	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("DEEPSEEK_API_KEY is not set")
	}

	transaction, _, err := llm.NewDeepSeekClient(apiKey).ExtractTransaction(rawText)
	if err != nil {
		return nil, err
	}
	return transaction, nil
}
