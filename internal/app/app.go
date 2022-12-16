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
	CompareFaces(source, target []byte) (int, error)
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

var (
	ErrRequest         = errors.New("request error")
	ErrDownload        = errors.New("unable to download a file")
	ErrServerNotExists = errors.New("remove server doesn't exist")
	ErrFileRead        = errors.New("unable to read a file")
)

func New(logger Logger, config Config, recognitionClient RecognitionClient) (*Application, error) {
	return &Application{
		Logger:            logger,
		Config:            config,
		RecognitionClient: recognitionClient,
	}, nil
}

func (app *Application) CompareImages(urls []string) (string, []string, []error) {

	imagesBytes := app.downloadImagesByUrls(urls)

	urlsCnt := len(urls)
	source := imagesBytes[0]
	targets := imagesBytes[1:]
	unmatched := make([]string, 0, urlsCnt)
	errs := make([]error, 0, urlsCnt)

	for _, target := range targets {
		unmatchedCnt, err := app.RecognitionClient.CompareFaces(source.bytes, target.bytes)

		if unmatchedCnt > 0 {
			unmatched = append(unmatched, target.url)
		}

		if err != nil {
			e := fmt.Errorf("unable to compare images %s and %s: %w", source.url, target.url, err)
			errs = append(errs, e)
		}
	}

	return source.url, unmatched, errs
}

func (app *Application) downloadImagesByUrls(urls []string) []ImagePair {

	imagePairs := make([]ImagePair, len(urls))
	for i, url := range urls {
		bytes, err := app.downloadByURL(url)
		if err == nil {
			imagePairs[i] = ImagePair{url, bytes}
		}
	}

	return imagePairs
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

	bytes, err := io.ReadAll(response.Body)
	if err != nil {
		return []byte{}, fmt.Errorf("%w: %s", ErrFileRead, err)
	}

	return bytes, nil
}
