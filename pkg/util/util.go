package util

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/nice-pink/goutil/pkg/log"
)

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
	return strings.ToLower(value) == "true"
}

func DownloadHttp(url string, filepath string) error {
	log.Info("http download:", url)

	out, err := os.Create(filepath)
	if err != nil {
		log.Err(err, "Could not create file.")
		return err
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		log.Err(err, "Could not request url.")
		return err
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		log.Error("bad status: %s", resp.Status)
		return errors.New("bad status: " + resp.Status)
	}

	n, err := io.Copy(out, resp.Body)
	if err != nil {
		log.Err(err, "Could not copy data to file.")
		return err
	}
	log.Info("Downloaded file with", n, "bytes")
	return nil
}

func UploadHttp(url string, filepath string, contentType string) error {
	log.Info("http upload", filepath, "to", url, "with content type", contentType)

	file, err := os.Open(filepath)
	if err != nil {
		log.Err(err, "Could not read file", filepath)
		return err
	}

	req, err := http.NewRequest(http.MethodPut, url, file)
	if err != nil {
		logPutError(err)
		return err
	}
	req.Header.Add("Content-Type", contentType)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		logSendError(err)
		return err
	}
	defer res.Body.Close()
	return nil
}

func PostRequest(url string, body []byte, headers map[string]string, printBody bool) error {
	log.Info("post", url)
	if printBody {
		log.Info(string(body))
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
		logPutError(err)
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
		logSendError(err)
		return err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	bodyString := string(bodyBytes)
	log.Info(bodyString)

	return nil
}

func GetRequest(url string, headers map[string]string) error {
	log.Info("get", url)

	// setup request
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		logPutError(err)
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
		logSendError(err)
		return err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	bodyString := string(bodyBytes)
	log.Info(bodyString)

	return nil
}

// errors

func logPutError(err error) {
	log.Err(err, "Could not create put request.")
}

func logSendError(err error) {
	log.Err(err, "Could not send request.")
}

// array

func RemoveFromStringArray(s []string, i int) []string {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}
