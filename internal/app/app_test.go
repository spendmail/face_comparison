package app

import (
	awsClient "github.com/spendmail/face_comparison/internal/aws"
	internalconfig "github.com/spendmail/face_comparison/internal/config"
	internallogger "github.com/spendmail/face_comparison/internal/logger"
	"github.com/stretchr/testify/require"
	_ "image/jpeg"
	"testing"
)

func TestApplication(t *testing.T) {
	t.Run("do test", func(t *testing.T) {
		config, err := internalconfig.New("../../configs/face_comparison.toml")
		require.NoError(t, err, "should be without errors")

		logger, err := internallogger.New(config)
		require.NoError(t, err, "should be without errors")

		recognitionClient, err := awsClient.NewRecognitionClient(config, logger)
		require.NoError(t, err, "should be without errors")

		app, err := New(logger, config, recognitionClient)
		require.NoError(t, err, "should be without errors")

		//urls := []string{"http://34.233.56.138/images/victor_man/82.jpeg", "http://34.233.56.138/images/victor_man/83.jpg", "http://34.233.56.138/images/victor_man/84.jpeg", "http://34.233.56.138/images/victor_man/85.jpg", "http://34.233.56.138/images/victor_man/86.jpg", "http://34.233.56.138/images/victor_man/87.jpg", "http://34.233.56.138/images/victor_man/88.jpeg", "http://34.233.56.138/images/victor_man/89.jpg", "http://34.233.56.138/images/victor_man/90.jpg", "http://34.233.56.138/images/victor_man/91.jpg"}
		urls := []string{"http://34.233.56.138/images/victor_man/82.jpeg", "http://34.233.56.138/images/victor_man/83.jpg", "http://34.233.56.138/images/victor_man/84.jpeg", "http://34.233.56.138/images/victor_man/85.jpg", "http://34.233.56.138/images/victor_man/86.jpg", "http://34.233.56.138/images/victor_man/87.jpg", "http://34.233.56.138/images/victor_man/88.jpeg", "http://34.233.56.138/images/victor_man/89.jpg", "http://34.233.56.138/images/victor_man/90.jpg", "http://34.233.56.138/images/victor_man/91.jpg"}
		//pairs, errs := app.downloadImagesByUrlsConcurrently(urls)
		//pairs, errs := app.downloadImagesByUrlsWithChannels(urls)
		//
		//require.Equal(t, len(errs), 0, "number of errors is greater that 0")
		//require.Equal(t, len(pairs), 10, "number of pairs is fewer that 10")

		_, _, _, _, _ = app.CompareImages(urls)

	})
}
