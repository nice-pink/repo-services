package util

import (
	"bytes"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
)

// env

func GetEnvString(key string, fallback string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	return value
}

func GetEnvBool(key string, fallback bool) bool {
	value, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	value = strings.ToLower(value)
	return value == "true"
}

// array

func RemoveFromStringArray(s []string, i int) []string {
	if i >= len(s) {
		return s
	}
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

// network

func DownloadHttp(url string, filepath string) error {
	slog.Default().Info("http_download", "url", url)

	out, err := os.Create(filepath)
	if err != nil {
		slog.Default().Error("http_download_create_file", "err", err)
		return err
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		slog.Default().Error("http_download_request", "err", err)
		return err
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		slog.Default().Error("http_download_bad_status", "status", resp.Status)
		return errors.New("bad status: " + resp.Status)
	}

	n, err := io.Copy(out, resp.Body)
	if err != nil {
		slog.Default().Error("http_download_copy", "err", err)
		return err
	}
	slog.Default().Info("http_download_ok", "bytes", n)
	return nil
}

func UploadHttp(url string, filepath string, contentType string) error {
	slog.Default().Info("http_upload", "filepath", filepath, "url", url, "contentType", contentType)

	file, err := os.Open(filepath)
	if err != nil {
		slog.Default().Error("http_upload_open", "err", err, "filepath", filepath)
		return err
	}

	req, err := http.NewRequest(http.MethodPut, url, file)
	if err != nil {
		slog.Default().Error("http_put_create", "err", err)
		return err
	}
	req.Header.Add("Content-Type", contentType)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		slog.Default().Error("http_put_send", "err", err)
		return err
	}
	defer res.Body.Close()
	return nil
}

func PostRequest(url string, body []byte, headers map[string]string, printBody bool) error {
	slog.Default().Info("http_post", "url", url)
	if printBody {
		slog.Default().Info("http_post_body", "body", string(body))
	}

	// setup request
	var req *http.Request
	var err error
	if len(body) == 0 {
		req, err = http.NewRequest(http.MethodPost, url, nil)
	} else {
		reader := bytes.NewReader(body)
		req, err = http.NewRequest(http.MethodPost, url, reader)
	}
	if err != nil {
		slog.Default().Error("http_post_create", "err", err)
		return err
	}

	// set additional headers
	for k, v := range headers {
		req.Header.Add(k, v)
	}

	// request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		slog.Default().Error("http_post_send", "err", err)
		return err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	bodyString := string(bodyBytes)
	slog.Default().Info("http_post_response", "body", bodyString)

	return nil
}

func GetRequest(url string, headers map[string]string) error {
	slog.Default().Info("http_get", "url", url)

	// setup request
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		slog.Default().Error("http_get_create", "err", err)
		return err
	}

	// set additional headers
	for k, v := range headers {
		req.Header.Add(k, v)
	}

	// request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		slog.Default().Error("http_get_send", "err", err)
		return err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	bodyString := string(bodyBytes)
	slog.Default().Info("http_get_response", "body", bodyString)

	return nil
}
