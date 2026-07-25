package httpapi

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	"parce/internal/model"
	"parce/internal/pipeline"
)

const maxUploadBytes = 20 << 20

func NewMux(apiKey string) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", handleHealth)
	mux.Handle("POST /parse", requireAPIKey(apiKey)(http.HandlerFunc(handleParse)))
	return mux
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func handleParse(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadBytes)

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "missing file field: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "read upload: "+err.Error(), http.StatusBadRequest)
		return
	}

	transaction, err := extract(data)
	if err != nil {
		slog.Error("extract transaction", "error", err, "filename", header.Filename)
		http.Error(w, "extract transaction: "+err.Error(), http.StatusUnprocessableEntity)
		return
	}

	result := model.Result{
		Transaction: *transaction,
		FilePath:    header.Filename,
		Timestamp:   time.Now().Format("2006-01-02_15-04-05"),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		slog.Error("encode response", "error", err)
	}
}

func extract(data []byte) (*model.Transaction, error) {
	if pipeline.IsPDF(data) {
		tmpPath, cleanup, err := writeTemp(data, "upload-*.pdf")
		if err != nil {
			return nil, err
		}
		defer cleanup()
		return pipeline.ExtractViaPDF(tmpPath)
	}

	transaction, err := pipeline.ExtractViaVision(data)
	if err == nil {
		return transaction, nil
	}
	slog.Warn("vision extraction failed, falling back to OCR", "error", err)

	tmpPath, cleanup, err := writeTemp(data, "upload-*.img")
	if err != nil {
		return nil, err
	}
	defer cleanup()
	return pipeline.ExtractViaOCR(tmpPath)
}

func writeTemp(data []byte, pattern string) (path string, cleanup func(), err error) {
	tmp, err := os.CreateTemp("", pattern)
	if err != nil {
		return "", nil, fmt.Errorf("create temp file: %w", err)
	}
	defer tmp.Close()

	if _, err := tmp.Write(data); err != nil {
		os.Remove(tmp.Name())
		return "", nil, fmt.Errorf("write temp file: %w", err)
	}
	return tmp.Name(), func() { os.Remove(tmp.Name()) }, nil
}
