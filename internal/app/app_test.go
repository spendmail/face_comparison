package app

import (
	"fmt"
	awsClient "github.com/spendmail/face_comparison/internal/aws"
	internalconfig "github.com/spendmail/face_comparison/internal/config"
	internallogger "github.com/spendmail/face_comparison/internal/logger"
	"github.com/stretchr/testify/require"
	_ "image/jpeg"
	"testing"
)

func TestApplication(t *testing.T) {
	t.Run("app initialisation", func(t *testing.T) {
		config, err := internalconfig.New("../../configs/face_comparison.toml")
		require.NoError(t, err, "should be without errors")

		logger, err := internallogger.New(config)
		require.NoError(t, err, "should be without errors")

		recognitionClient, err := awsClient.NewRecognitionClient(config, logger)
		require.NoError(t, err, "should be without errors")

		_, err = New(logger, config, recognitionClient)
		require.NoError(t, err, "should be without errors")
	})

	t.Run("10 matches", func(t *testing.T) {
		config, _ := internalconfig.New("../../configs/face_comparison.toml")
		logger, _ := internallogger.New(config)
		recognitionClient, _ := awsClient.NewRecognitionClient(config, logger)
		app, _ := New(logger, config, recognitionClient)

		urls := []string{"http://34.233.56.138/images/victor_man/82.jpeg", "http://34.233.56.138/images/victor_man/83.jpg", "http://34.233.56.138/images/victor_man/84.jpeg", "http://34.233.56.138/images/victor_man/85.jpg", "http://34.233.56.138/images/victor_man/86.jpg", "http://34.233.56.138/images/victor_man/87.jpg", "http://34.233.56.138/images/victor_man/88.jpeg", "http://34.233.56.138/images/victor_man/89.jpg", "http://34.233.56.138/images/victor_man/90.jpg", "http://34.233.56.138/images/victor_man/91.jpg"}
		_, unmatched, multipleFaces, facesNotFound, _, errs := app.CompareImages(urls)
		require.Equal(t, 0, len(unmatched), "unmatched != 0")
		require.Equal(t, 0, len(multipleFaces), "multipleFaces != 0")
		require.Equal(t, 0, len(facesNotFound), "facesNotFound != 0")
		require.Equal(t, 0, len(errs), "errs != 0")
	})

	t.Run("unsupported file type", func(t *testing.T) {
		config, _ := internalconfig.New("../../configs/face_comparison.toml")
		logger, _ := internallogger.New(config)
		recognitionClient, _ := awsClient.NewRecognitionClient(config, logger)
		app, _ := New(logger, config, recognitionClient)

		urls := []string{"http://34.233.56.138/images/victor_man/82.jpeg", "http://34.233.56.138/images/unsupported/IMG_0004.HEIC"}
		_, _, _, _, _, errs := app.CompareImages(urls)
		require.Equal(t, 2, len(errs), "errs != 0")
	})

	t.Run("multiple faces", func(t *testing.T) {
		config, _ := internalconfig.New("../../configs/face_comparison.toml")
		logger, _ := internallogger.New(config)
		recognitionClient, _ := awsClient.NewRecognitionClient(config, logger)
		app, _ := New(logger, config, recognitionClient)

		urls := []string{"http://34.233.56.138/images/victor_man/82.jpeg", "http://34.233.56.138/images/unsupported/multiple_faces_1.jpeg", "http://34.233.56.138/images/unsupported/multiple_faces_2.jpeg", "http://34.233.56.138/images/unsupported/multiple_faces_3.jpeg"}
		_, _, multipleFaces, _, _, _ := app.CompareImages(urls)
		require.True(t, len(multipleFaces) >= 2 && len(multipleFaces) <= 3, fmt.Sprintf("multipleFaces != 2 or 3, %d given", len(multipleFaces)))
	})

	t.Run("faces not found", func(t *testing.T) {
		config, _ := internalconfig.New("../../configs/face_comparison.toml")
		logger, _ := internallogger.New(config)
		recognitionClient, _ := awsClient.NewRecognitionClient(config, logger)
		app, _ := New(logger, config, recognitionClient)

		urls := []string{"http://34.233.56.138/images/victor_man/82.jpeg", "http://34.233.56.138/images/unsupported/no_faces.jpg"}
		_, _, _, facesNotFound, _, _ := app.CompareImages(urls)
		require.Equal(t, 1, len(facesNotFound), fmt.Sprintf("facesNotFound != 0, %d given", len(facesNotFound)))
	})

	t.Run("faces not found", func(t *testing.T) {
		config, _ := internalconfig.New("../../configs/face_comparison.toml")
		logger, _ := internallogger.New(config)
		recognitionClient, _ := awsClient.NewRecognitionClient(config, logger)
		app, _ := New(logger, config, recognitionClient)

		urls := []string{"http://34.233.56.138/images/victor_man/82.jpeg", "http://34.233.56.138/images/celebahq_identity_10111_woman/15006.jpg", "http://34.233.56.138/images/celebahq_identity_8190_man/1269.jpg", "http://34.233.56.138/images/dicaprio_man/32.jpg", "http://34.233.56.138/images/mlexandra_woman/62.jpg", "http://34.233.56.138/images/sergey_man/72.jpg", "http://34.233.56.138/images/angelina_jolie_woman/10.jpeg", "http://34.233.56.138/images/celebahq_identity_5046_woman/15277.jpg", "http://34.233.56.138/images/celebahq_identity_8960_man/10944.jpg", "http://34.233.56.138/images/emma_stone_woman/42.jpeg", "http://34.233.56.138/images/anya_woman/12.jpeg", "http://34.233.56.138/images/celebahq_identity_8189_woman/16399.jpg", "http://34.233.56.138/images/cumberbatch_man/22.jpg", "http://34.233.56.138/images/kate_woman/52.jpg", "http://34.233.56.138/images/victor_man/91.jpg"}
		_, unmatched, _, _, _, _ := app.CompareImages(urls)
		require.True(t, len(unmatched) >= 13 && len(unmatched) <= 14, fmt.Sprintf("unmatched != 13 or 14, %d given", len(unmatched)))
	})

	t.Run("gender male", func(t *testing.T) {
		config, _ := internalconfig.New("../../configs/face_comparison.toml")
		logger, _ := internallogger.New(config)
		recognitionClient, _ := awsClient.NewRecognitionClient(config, logger)
		app, _ := New(logger, config, recognitionClient)

		urls := []string{"http://34.233.56.138/images/victor_man/82.jpeg", "http://34.233.56.138/images/victor_man/83.jpg"}
		_, _, _, _, gender, _ := app.CompareImages(urls)
		require.Equal(t, "male", gender, "gander != male")
	})

	t.Run("gender female", func(t *testing.T) {
		config, _ := internalconfig.New("../../configs/face_comparison.toml")
		logger, _ := internallogger.New(config)
		recognitionClient, _ := awsClient.NewRecognitionClient(config, logger)
		app, _ := New(logger, config, recognitionClient)

		urls := []string{"http://34.233.56.138/images/emma_stone_woman/42.jpeg", "http://34.233.56.138/images/emma_stone_woman/43.jpg"}
		_, _, _, _, gender, _ := app.CompareImages(urls)
		require.Equal(t, "female", gender, "gander != female")
	})
}
