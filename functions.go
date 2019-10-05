package main

import (
	"html/template"
	"io/ioutil"
	"mime/multipart"
	"os"
)

func UploadImages(file multipart.File) (string, error) {
	tempFile, err := ioutil.TempFile("upload-images", "upload-*.png")
	if err != nil {
		return "", err
	}
	defer tempFile.Close()

	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		return "", err
	}

	_, err = tempFile.Write(fileBytes)
	if err != nil {
		return "", err
	}
	return tempFile.Name(), nil
}

func DeleteImages(src string) error {
	err := os.Remove(src)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func safeSocial(url string) template.URL {
	return template.URL(url)
}
