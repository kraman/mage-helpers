package net

import (
	"io"
	"path/filepath"
	"net/http"
	"os"
)

func Download(url, dest string) error {
	if _, err := os.Stat(dest); err == nil {
		return nil
	}

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := os.MkdirAll(filepath.Dir(dest), 0750); err != nil {
		return err
	}

	out, err := os.Create(filepath.Join(dest))
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err = io.Copy(out, resp.Body); err != nil {
		return err
	}
	return nil
}