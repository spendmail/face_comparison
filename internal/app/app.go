package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
)

type Logger interface {
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
}

type Config interface {
}

type RecognitionClient interface {
	CompareFaces(source, target []byte) (int, int, error)
}

type Application struct {
	Logger            Logger
	Config            Config
	RecognitionClient RecognitionClient
}

type ImagePair struct {
	url   string
	bytes []byte
}

const (
	MimePng  = "image/png"
	MimeJpeg = "image/jpeg"
)

var (
	ErrRequest          = errors.New("request error")
	ErrDownload         = errors.New("unable to download a file")
	ErrServerNotExists  = errors.New("remove server doesn't exist")
	ErrFileRead         = errors.New("unable to read a file")
	ErrFileNotSupported = errors.New("unsupported file type")
	ErrNotEnoughImage   = errors.New("not enough images to compare")
)

func New(logger Logger, config Config, recognitionClient RecognitionClient) (*Application, error) {
	return &Application{
		Logger:            logger,
		Config:            config,
		RecognitionClient: recognitionClient,
	}, nil
}

func (app *Application) CompareImages(urls []string) (string, []string, []string, []string, []error) {

	urlsCnt := len(urls)

	// not enough photos
	if urlsCnt < 2 {
		return "", []string{}, []string{}, []string{}, []error{ErrNotEnoughImage}
	}

	unmatched := make([]string, 0, urlsCnt)
	multipleFaces := make([]string, 0, urlsCnt)
	facesNotFound := make([]string, 0, urlsCnt)

	// downloading images
	imagesBytes, errs := app.downloadImagesByUrls(urls)

	// not enough photos after filtering
	if len(imagesBytes) < 2 {
		errs = append(errs, fmt.Errorf("%w: some of the images were probably filtered", ErrNotEnoughImage))
		return "", []string{}, []string{}, []string{}, errs
	}

	source := imagesBytes[0]
	targets := imagesBytes[1:]

	// faces comparison
	for _, target := range targets {
		unmatchedCnt, matchedCnt, err := app.RecognitionClient.CompareFaces(source.bytes, target.bytes)

		if unmatchedCnt == 1 {
			unmatched = append(unmatched, target.url)
		} else if unmatchedCnt > 1 {
			multipleFaces = append(multipleFaces, target.url)
		}

		if unmatchedCnt == 0 && matchedCnt == 0 {
			facesNotFound = append(facesNotFound, target.url)
		}

		if err != nil {
			e := fmt.Errorf("unable to compare images %s and %s: %w", source.url, target.url, err)
			errs = append(errs, e)
		}
	}

	return source.url, unmatched, multipleFaces, facesNotFound, errs
}

func (app *Application) downloadImagesByUrls(urls []string) ([]ImagePair, []error) {

	imagePairs := make([]ImagePair, 0, len(urls))
	errs := make([]error, 0, len(imagePairs))

	for _, url := range urls {
		imageBytes, err := app.downloadByURL(url)

		if err != nil {
			errs = append(errs, err)
			continue
		}

		err = app.extensionValidate(imageBytes)
		if err != nil {
			errs = append(errs, fmt.Errorf("%w: %s", err, url))
			continue
		}

		imagePairs = append(imagePairs, ImagePair{url, imageBytes})
	}

	return imagePairs, errs
}

func (app *Application) downloadByURL(url string) ([]byte, error) {
	request, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		return []byte{}, fmt.Errorf("%w: %s", ErrRequest, err)
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		// Identifying wrong domain name errors.
		var DNSError *net.DNSError
		if errors.As(err, &DNSError) {
			return []byte{}, fmt.Errorf("%w: %s", ErrServerNotExists, err)
		}

		return []byte{}, fmt.Errorf("%w: %s", ErrDownload, err)
	}
	defer response.Body.Close()

	responseBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return []byte{}, fmt.Errorf("%w: %s", ErrFileRead, err)
	}

	return responseBytes, nil
}

func (app *Application) extensionValidate(imageBytes []byte) error {

	mimeType := http.DetectContentType(imageBytes)

	if mimeType != MimePng && mimeType != MimeJpeg {
		return ErrFileNotSupported
	}

	return nil
}
