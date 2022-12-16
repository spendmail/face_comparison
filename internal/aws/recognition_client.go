package s3

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rekognition"
)

type Config interface {
	GetAccessKeyId() string
	GetSecretAccessKey() string
	GetRegion() string
	GetSimilarityThreshold() float64
}

type Logger interface {
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
}

type Client struct {
	config Config
	logger Logger
	svc    *rekognition.Rekognition
}

func NewRecognitionClient(config Config, logger Logger) (*Client, error) {

	awsConfig := aws.NewConfig()
	awsConfig.WithCredentials(credentials.NewStaticCredentials(config.GetAccessKeyId(), config.GetSecretAccessKey(), ""))
	awsConfig.WithRegion(config.GetRegion())

	return &Client{
		config: config,
		logger: logger,
		svc:    rekognition.New(session.New(), awsConfig),
	}, nil
}

func (c *Client) CompareFaces(source, target []byte) (int, error) {

	input := &rekognition.CompareFacesInput{
		SimilarityThreshold: aws.Float64(c.config.GetSimilarityThreshold()),
		SourceImage: &rekognition.Image{
			Bytes: source,
		},
		TargetImage: &rekognition.Image{
			Bytes: target,
		},
	}

	result, err := c.svc.CompareFaces(input)
	unmatchedFacesCnt := len(result.UnmatchedFaces)

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case rekognition.ErrCodeInvalidParameterException:
				return unmatchedFacesCnt, fmt.Errorf("%s: %w", rekognition.ErrCodeInvalidParameterException, aerr)
			case rekognition.ErrCodeInvalidS3ObjectException:
				return unmatchedFacesCnt, fmt.Errorf("%s: %w", rekognition.ErrCodeInvalidS3ObjectException, aerr)
			case rekognition.ErrCodeImageTooLargeException:
				return unmatchedFacesCnt, fmt.Errorf("%s: %w", rekognition.ErrCodeImageTooLargeException, aerr)
			case rekognition.ErrCodeAccessDeniedException:
				return unmatchedFacesCnt, fmt.Errorf("%s: %w", rekognition.ErrCodeAccessDeniedException, aerr)
			case rekognition.ErrCodeInternalServerError:
				return unmatchedFacesCnt, fmt.Errorf("%s: %w", rekognition.ErrCodeInternalServerError, aerr)
			case rekognition.ErrCodeThrottlingException:
				return unmatchedFacesCnt, fmt.Errorf("%s: %w", rekognition.ErrCodeThrottlingException, aerr)
			case rekognition.ErrCodeProvisionedThroughputExceededException:
				return unmatchedFacesCnt, fmt.Errorf("%s: %w", rekognition.ErrCodeProvisionedThroughputExceededException, aerr)
			case rekognition.ErrCodeInvalidImageFormatException:
				return unmatchedFacesCnt, fmt.Errorf("%s: %w", rekognition.ErrCodeInvalidImageFormatException, aerr)
			default:
				return unmatchedFacesCnt, fmt.Errorf("compare faces error: %w", aerr)
			}
		} else {
			return unmatchedFacesCnt, fmt.Errorf("compare faces error: %w", err)
		}
	}

	return unmatchedFacesCnt, nil
}
