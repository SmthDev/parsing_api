package model

type Transaction struct {
	CheckNumber string  `json:"check_number,omitempty"`
	Date        string  `json:"date,omitempty"`
	Bank        string  `json:"bank,omitempty"`
	Card        string  `json:"card,omitempty"`
	Payer       string  `json:"payer,omitempty"`
	Service     string  `json:"service,omitempty"`
	Account     string  `json:"account,omitempty"`
	Recipient   string  `json:"recipient,omitempty"`
	Amount      float64 `json:"amount,omitempty"`
	Currency    string  `json:"currency,omitempty"`
	Test        string  `json:"test,omitempty"`
}

type Result struct {
	Transaction
	FilePath  string `json:"file_path"`
	Timestamp string `json:"timestamp"`
}
