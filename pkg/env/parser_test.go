package env

import (
	"os"
	"testing"
	"time"
)

type testHTTPConfig struct {
	Host        string        `env:"HOST"`
	Port        int           `env:"PORT,8080"`
	Timeout     time.Duration `env:"TIMEOUT,5s"`
	IdleTimeout time.Duration `env:"IDLE_TIMEOUT"`
}

type testParamsConfig struct {
	UintValue   uint   `env:"VALUE_1"`
	UintValue8  uint8  `env:"VALUE_8"`
	UintValue16 uint16 `env:"VALUE_16"`
	UintValue32 uint32 `env:"VALUE_32"`
	UintValue64 uint64 `env:"VALUE_64"`
	IntValue    int    `env:"INT_VALUE_1"`
	IntValue8   int8   `env:"INT_VALUE_8"`
	IntValue16  int16  `env:"INT_VALUE_16"`
	IntValue32  int32  `env:"INT_VALUE_32"`
	IntValue64  int64  `env:"INT_VALUE_64"`
}

type testCalculationsConfig struct {
	Percent    float64          `env:"PERCENT"`
	PercentLow float32          `env:"PERCENT_LOW"`
	Params     testParamsConfig `env:"PARAMS"`
}

type testConfig struct {
	APIKey      string `env:"API_KEY"`
	EnableCache bool   `env:"ENABLE_CACHE"`

	HTTP testHTTPConfig `env:"HTTP"`

	Calculations testCalculationsConfig `env:"CALCULATIONS"`

	Interval1 time.Duration `env:"INTERVAL,10m"`
	Interval2 time.Duration `env:"INTERVAL_2,1h"`
}

func TestParse(t *testing.T) {
	testCases := []struct {
		prepareEnv    func()
		expectedValue testConfig
	}{
		{
			prepareEnv: func() {
				os.Setenv("API_KEY", "test-key")
				os.Setenv("ENABLE_CACHE", "true")

				os.Setenv("HTTP_HOST", "localhost")
				os.Setenv("HTTP_IDLE_TIMEOUT", "5m")

				os.Setenv("CALCULATIONS_PERCENT", "11")
				os.Setenv("CALCULATIONS_PERCENT_LOW", "0.1488")
				os.Setenv("CALCULATIONS_PARAMS_VALUE_1", "1")
				os.Setenv("CALCULATIONS_PARAMS_VALUE_8", "8")
				os.Setenv("CALCULATIONS_PARAMS_VALUE_16", "16")
				os.Setenv("CALCULATIONS_PARAMS_VALUE_32", "32")
				os.Setenv("CALCULATIONS_PARAMS_VALUE_64", "64")
				os.Setenv("CALCULATIONS_PARAMS_INT_VALUE_1", "11")
				os.Setenv("CALCULATIONS_PARAMS_INT_VALUE_8", "88")
				os.Setenv("CALCULATIONS_PARAMS_INT_VALUE_16", "166")
				os.Setenv("CALCULATIONS_PARAMS_INT_VALUE_32", "322")
				os.Setenv("CALCULATIONS_PARAMS_INT_VALUE_64", "644")
			},
			expectedValue: testConfig{
				APIKey:      "test-key",
				EnableCache: true,

				HTTP: testHTTPConfig{
					Host:        "localhost",
					Port:        8080,
					Timeout:     5 * time.Second,
					IdleTimeout: 5 * time.Minute,
				},

				Calculations: testCalculationsConfig{
					Percent:    11,
					PercentLow: 0.1488,
					Params: testParamsConfig{
						UintValue:   1,
						UintValue8:  8,
						UintValue16: 16,
						UintValue32: 32,
						UintValue64: 64,
						IntValue:    11,
						IntValue8:   88,
						IntValue16:  166,
						IntValue32:  322,
						IntValue64:  644,
					},
				},

				Interval1: 10 * time.Minute,
				Interval2: time.Hour,
			},
		},
	}

	for _, tc := range testCases {
		os.Clearenv()
		tc.prepareEnv()
		cfg, err := Parse[testConfig]()
		if err != nil {
			t.Error(err)
			continue
		}

		if cfg != tc.expectedValue {
			t.Errorf("expected value %v; actual value: %v", tc.expectedValue, cfg)
		}
	}
}
