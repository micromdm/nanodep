package client

import (
	"context"
	"testing"
)

func TestDefaultConfigRetreiver(t *testing.T) {
	for _, cfg := range []struct {
		name        string
		cfg         *Config
		expectedCfg *Config
	}{
		{
			name:        "nil",
			cfg:         nil,
			expectedCfg: &Config{BaseURL: DefaultBaseURL},
		},
		{
			name:        "empty",
			cfg:         &Config{},
			expectedCfg: &Config{BaseURL: DefaultBaseURL},
		},
		{
			name:        "existent",
			cfg:         &Config{BaseURL: "foo"},
			expectedCfg: &Config{BaseURL: "foo"},
		},
	} {
		t.Run(cfg.name, func(t *testing.T) {
			c := NewDefaultConfigRetreiver(cfgRetriever{cfg: cfg.cfg})
			actualCfg, err := c.RetrieveConfig(context.Background(), "foo")
			if err != nil {
				t.Fatal(err)
			}
			if actualCfg == nil {
				t.Fatal("expected not-nil config")
			}
			if *actualCfg != *cfg.expectedCfg {
				t.Fatalf("unexpected base URL: %+v vs %+v", actualCfg, cfg.expectedCfg)
			}
		})
	}

}

type cfgRetriever struct {
	cfg *Config
}

func (c cfgRetriever) RetrieveConfig(context.Context, string) (*Config, error) {
	return c.cfg, nil
}
