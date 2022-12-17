package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
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
	PredictGender(source []byte) (string, error)
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
	ErrServerNotExists  = errors.New("remote server doesn't exist")
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

func (app *Application) CompareImages(urls []string) (string, []string, []string, []string, string, []error) {

	urlsCnt := len(urls)

	// not enough photos
	if urlsCnt < 2 {
		return "", []string{}, []string{}, []string{}, "", []error{ErrNotEnoughImage}
	}

	unmatched := make([]string, 0, urlsCnt)
	multipleFaces := make([]string, 0, urlsCnt)
	facesNotFound := make([]string, 0, urlsCnt)

	// downloading images
	//imagesBytes, errs := app.downloadImagesByUrls(urls)
	imagesBytes, errs := app.downloadImagesByUrlsWithChannels(urls)

	// not enough photos after filtering
	if len(imagesBytes) < 2 {
		errs = append(errs, fmt.Errorf("%w: some of the images were probably filtered", ErrNotEnoughImage))
		return "", []string{}, []string{}, []string{}, "", errs
	}

	source := imagesBytes[0]
	targets := imagesBytes[1:]

	cnt := len(targets)
	unmatchedChan := make(chan string, cnt)
	multipleFacesChan := make(chan string, cnt)
	facesNotFoundChan := make(chan string, cnt)
	errsChan := make(chan error, cnt)

	wg := sync.WaitGroup{}

	// faces comparison
	for _, target := range targets {
		wg.Add(1)
		go func(p ImagePair) {
			defer wg.Done()
			unmatchedCnt, matchedCnt, err := app.RecognitionClient.CompareFaces(source.bytes, p.bytes)

			if unmatchedCnt == 1 {
				unmatchedChan <- p.url
			} else if unmatchedCnt > 1 {
				multipleFacesChan <- p.url
			}

			if unmatchedCnt == 0 && matchedCnt == 0 {
				facesNotFoundChan <- p.url
			}

			if err != nil {
				e := fmt.Errorf("unable to compare images %s and %s: %w", source.url, p.url, err)
				errsChan <- e
			}
		}(target)
	}

	wg.Wait()

	close(unmatchedChan)
	close(multipleFacesChan)
	close(facesNotFoundChan)
	close(errsChan)

	for {
		val, ok := <-unmatchedChan
		if !ok {
			break
		}
		unmatched = append(unmatched, val)
	}

	for {
		val, ok := <-multipleFacesChan
		if !ok {
			break
		}
		multipleFaces = append(multipleFaces, val)
	}

	for {
		val, ok := <-facesNotFoundChan
		if !ok {
			break
		}
		facesNotFound = append(facesNotFound, val)
	}

	for {
		val, ok := <-errsChan
		if !ok {
			break
		}
		errs = append(errs, val)
	}

	gender, err := app.RecognitionClient.PredictGender(source.bytes)
	if err != nil {
		errs = append(errs, fmt.Errorf("%s: %w", source.url, err))
	}

	return source.url, unmatched, multipleFaces, facesNotFound, gender, errs
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

func (app *Application) downloadImagesByUrlsWithChannels(urls []string) ([]ImagePair, []error) {

	urlsLen := len(urls)
	imagePairs := make([]ImagePair, 0, urlsLen)
	errs := make([]error, 0, urlsLen)
	var wg sync.WaitGroup

	errsChan := make(chan error, urlsLen)
	pairsChan := make(chan ImagePair, urlsLen)

	for _, url := range urls {

		wg.Add(1)
		go func(url string) {
			defer wg.Done()

			imageBytes, err := app.downloadByURL(url)

			if err != nil {
				errsChan <- err
				return
			}

			err = app.extensionValidate(imageBytes)
			if err != nil {
				errsChan <- fmt.Errorf("%w: %s", err, url)
				return
			}

			pairsChan <- ImagePair{url, imageBytes}
		}(url)
	}

	wg.Wait()
	close(errsChan)
	close(pairsChan)

	for {
		e, ok := <-errsChan
		if !ok {
			break
		}
		errs = append(errs, e)
	}

	for {
		pair, ok := <-pairsChan
		if !ok {
			break
		}
		imagePairs = append(imagePairs, pair)
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
