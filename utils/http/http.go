package http

import (
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

func GetPage(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func DownloadFile(url string, filepath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	file, createFileErr := os.Create(filepath)
	if createFileErr != nil {
		return createFileErr
	}
	defer file.Close()
	_, copyErr := io.Copy(file, resp.Body)
	if copyErr != nil {
		return copyErr
	}
	return nil
}
