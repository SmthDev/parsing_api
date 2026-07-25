package llm

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"parce/internal/model"
)

const defaultGroqModel = "qwen/qwen3.6-27b"

const groqVisionPrompt = `Ты — парсер банковских карт-чеков (чеков об оплате/переводе). Тебе дано фото такого чека.
Извлеки данные о транзакции. Верни ТОЛЬКО валидный JSON без пояснений и markdown-блоков, без лишних полей.
Формат:
{"check_number": "номер карт-чека", "date": "дата и время платежа", "bank": "банк-отправитель платежа", "card": "номер карты (маскированный)", "payer": "плательщик", "service": "услуга/назначение платежа", "account": "лицевой счёт или № счёта получателя", "recipient": "получатель платежа", "amount": сумма (число), "currency": "валюта"}
Если поле не видно на чеке — оставь пустую строку (для amount — 0).
Сумма должна быть точной, как на чеке, без пробелов и с точкой в качестве разделителя.`

type GroqClient struct {
	apiKey string
}

func NewGroqClient(apiKey string) *GroqClient {
	return &GroqClient{apiKey: apiKey}
}

func (c *GroqClient) ExtractTransactionFromImage(imageData []byte, mimeType string) (*model.Transaction, []byte, error) {
	b64 := base64.StdEncoding.EncodeToString(imageData)
	dataURL := fmt.Sprintf("data:%s;base64,%s", mimeType, b64)

	groqModel := os.Getenv("GROQ_MODEL")
	if groqModel == "" {
		groqModel = defaultGroqModel
	}

	requestBody := map[string]any{
		"model": groqModel,
		"messages": []map[string]any{
			{
				"role": "user",
				"content": []map[string]any{
					{
						"type": "text",
						"text": groqVisionPrompt,
					},
					{
						"type": "image_url",
						"image_url": map[string]any{
							"url": dataURL,
						},
					},
				},
			},
		},
		"response_format":       map[string]any{"type": "json_object"},
		"temperature":           0.2,
		"max_completion_tokens": 4096,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, nil, err
	}

	req, err := http.NewRequest("POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, bodyBytes, fmt.Errorf("Groq API error %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var apiResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(bodyBytes, &apiResp); err != nil {
		return nil, bodyBytes, err
	}
	if len(apiResp.Choices) == 0 {
		return nil, bodyBytes, fmt.Errorf("no response from Groq API")
	}

	content := strings.TrimSpace(apiResp.Choices[0].Message.Content)

	var transaction model.Transaction
	if err := json.Unmarshal([]byte(content), &transaction); err != nil {
		return nil, bodyBytes, fmt.Errorf("failed to parse Groq JSON: %w\nraw: %s", err, content)
	}
	if transaction.Amount == 0 && transaction.Bank == "" {
		return nil, bodyBytes, fmt.Errorf("Groq returned empty transaction\nraw: %s", content)
	}

	return &transaction, bodyBytes, nil
}
