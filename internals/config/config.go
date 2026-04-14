package config

import (
	"os"
	"strings"

	"github.com/DeepanshuChaid/Cogito-Ai.git/pkg/types"
)

func MustLoad(path string) *types.Config {
	config := &types.Config{
		Enabled: true,
		CompressOutput: false,
		Debug: false,
	}

	if path != "" {
		content, err := os.ReadFile(path)

		if err == nil {
			lines := strings.Split(string(content), "\n")
			for _, line :=  range lines {
				parts := strings.SplitN(line, "=", 2)
				if len(parts) == 2 {
					key := strings.TrimSpace(parts[0])
					val := strings.TrimSpace(parts[1])

					switch key {
					case "enabled":
						// IF THE CONDITION MATCHES THE ENABLED VALUE BECOME TRUE VICE VERSA
						config.Enabled = val == "true"
					case "context_file":
						config.ContextFile = val
					case "compress_output":
						config.CompressOutput = val == "true"
					}
				}
			}
		}
	}

	return config
}
