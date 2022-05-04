package config_test

import (
	"os"
	"testing"

	"code.com/config"
	"github.com/google/go-cmp/cmp"
)

func TestParse(t *testing.T) {
	expectedConf := config.Config{
		CronConfigs: config.CronConfigs{
			InventoryCron: config.CronConfig{
				Schedule:    "30 0 * * *",
				Description: "Cron to calculate inventory stats",
				Disabled:    false,
				NotifyEmail: []string{"jdoe@gmail.com"},
			},
			InvoicesCron: config.CronConfig{
				Schedule:    "10 0 * * *",
				Description: "Cron to generate invoices",
				Disabled:    true,
			},
		},
	}

	os.Args[1] = "-cron-config-file=testdata/cron_config.test.json"
	c, err := config.Parse()
	if err != nil {
		t.Fatalf("failed to parse config. %v", err)
	}

	if diff := cmp.Diff(expectedConf, c); diff != "" {
		t.Errorf("Configs are different (-want +got):\n%s", diff)
	}
}
