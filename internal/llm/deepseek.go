package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"parce/internal/model"
)

const apiURL = "https://api.deepseek.com/chat/completions"

const systemPrompt = `Ты — парсер банковских карт-чеков (чеков об оплате/переводе). Тебе дан результат OCR-распознавания такого чека.
Извлеки данные о транзакции и верни ТОЛЬКО валидный JSON без пояснений и markdown-блоков, без лишних полей.
Формат:
{"check_number": "номер карт-чека", "date": "дата и время платежа", "bank": "банк-отправитель платежа", "card": "номер карты (маскированный)", "payer": "плательщик", "service": "услуга/назначение платежа", "account": "лицевой счёт или № счёта получателя", "recipient": "получатель платежа", "amount": сумма (число), "currency": "валюта","test": "КОД УСЛУГИ ЕРИП"}
Если поле не видно в тексте — оставь пустую строку (для amount — 0).`

type DeepSeekClient struct {
	apiKey string
}

func NewDeepSeekClient(apiKey string) *DeepSeekClient {
	return &DeepSeekClient{apiKey: apiKey}
}

// ExtractTransaction calls the DeepSeek API and returns the parsed transaction plus the raw API response.
func (c *DeepSeekClient) ExtractTransaction(rawText string) (*model.Transaction, []byte, error) {
	requestBody := map[string]any{
		"model": "deepseek-v4-flash",
		"messages": []map[string]any{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": rawText},
		},
		"response_format": map[string]any{"type": "json_object"},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, nil, err
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewReader(jsonBody))
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
		return nil, bodyBytes, fmt.Errorf("DeepSeek API error %d: %s", resp.StatusCode, string(bodyBytes))
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
		return nil, bodyBytes, fmt.Errorf("no response from API")
	}

	content := strings.TrimSpace(apiResp.Choices[0].Message.Content)

	var transaction model.Transaction
	if err := json.Unmarshal([]byte(content), &transaction); err != nil {
		return nil, bodyBytes, fmt.Errorf("failed to parse transaction JSON: %w\nraw: %s", err, content)
	}
	if transaction.Amount == 0 && transaction.Bank == "" {
		return nil, bodyBytes, fmt.Errorf("DeepSeek returned empty transaction\nraw: %s", content)
	}
	return &transaction, bodyBytes, nil
}
