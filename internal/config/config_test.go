package config

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func resetFlagForTesting() {
	flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)
	flag.CommandLine.Usage = nil // Убираем стандартный Usage
}

func Test_Config(t *testing.T) {
	tests := []struct {
		name string
		args []string
		envs map[string]string
		want Config
	}{
		{
			name: "empty args and envs",
			args: []string{"cmd"},
			envs: map[string]string{
				"SERVER_ADDRESS": "",
				"BASE_URL":       "",
			},
			want: Config{
				ServerAddr: ":8080",
				BaseURL:    "http://localhost:8080",
			},
		},
		{
			name: "empty args, set envs",
			args: []string{"cmd"},
			envs: map[string]string{
				"SERVER_ADDRESS": "localhost:1085",
				"BASE_URL":       "http://127.0.0.1:1085",
			},
			want: Config{
				ServerAddr: "localhost:1085",
				BaseURL:    "http://127.0.0.1:1085",
			},
		},
		{
			name: "set args, set envs",
			args: []string{"cmd", "-a", ":443", "-b", "https://localhost:443"},
			envs: map[string]string{
				"SERVER_ADDRESS": "localhost:1085",
				"BASE_URL":       "http://127.0.0.1:1085",
			},
			want: Config{
				ServerAddr: "localhost:1085",
				BaseURL:    "http://127.0.0.1:1085",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// save original values
			oldArgs := os.Args
			oldEnvs := make(map[string]string)
			for k := range tt.envs {
				oldEnvs[k] = os.Getenv(k)
			}
			// restore original values
			defer func() {
				os.Args = oldArgs
				for k, v := range oldEnvs {
					os.Setenv(k, v)
				}
			}()

			os.Args = tt.args // set args
			for k, v := range tt.envs {
				os.Setenv(k, v) // set envs
			}

			resetFlagForTesting()
			k := GetConfig()

			assert.Equal(t, tt.want.ServerAddr, k.ServerAddr)
			assert.Equal(t, tt.want.BaseURL, k.BaseURL)
		})
	}
}
