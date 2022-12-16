package logger

import (
	"io/ioutil"
	"os"
	"testing"

	internalconfig "github.com/spendmail/face_comparison/internal/config"
	"github.com/stretchr/testify/require"
)

func TestLogger(t *testing.T) {
	t.Run("logger", func(t *testing.T) {
		config, err := internalconfig.New("../../configs/face_comparison.toml")
		if err != nil {
			t.Fatal(err)
		}

		f, err := os.CreateTemp("/tmp/", "")
		if err != nil {
			t.Fatal(err)
		}

		config.Logger.Level = "debug"
		config.Logger.File = f.Name()

		logger, err := New(config)
		if err != nil {
			t.Fatal(err)
		}

		logger.Debug("debug_message")
		logger.Info("info_message")
		logger.Warn("warn_message")
		logger.Error("error_message")

		b, err := ioutil.ReadFile(f.Name())
		if err != nil {
			t.Fatal(err)
		}

		content := string(b)

		require.Contains(t, content, "debug_message", "Log doesn't contain string error")
		require.Contains(t, content, "info_message", "Log doesn't contain string error")
		require.Contains(t, content, "warn_message", "Log doesn't contain string error")
		require.Contains(t, content, "error_message", "Log doesn't contain string error")
	})
}
