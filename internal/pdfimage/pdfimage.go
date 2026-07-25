package pdfimage

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)


func RenderFirstPage(pdfPath string) ([]byte, error) {
	bin, err := exec.LookPath("pdftoppm")
	if err != nil {
		return nil, fmt.Errorf("pdftoppm not found (install poppler-utils): %w", err)
	}

	tmpDir, err := os.MkdirTemp("", "pdfrender")
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	outPrefix := filepath.Join(tmpDir, "page")
	cmd := exec.Command(bin, "-png", "-r", "200", "-f", "1", "-l", "1", "-singlefile", pdfPath, outPrefix)
	if out, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("pdftoppm: %w: %s", err, out)
	}

	data, err := os.ReadFile(outPrefix + ".png")
	if err != nil {
		return nil, fmt.Errorf("read rendered page: %w", err)
	}
	return data, nil
}
