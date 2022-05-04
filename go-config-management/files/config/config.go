package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
)

type Config struct {
	// ...
	CronConfigs CronConfigs
}

type CronConfigs struct {
	InventoryCron CronConfig `json:"inventoryCron"`
	InvoicesCron  CronConfig `json:"invoicesCron"`
}

type CronConfig struct {
	Schedule    string   `json:"schedule"`
	Description string   `json:"desc"`
	Disabled    bool     `json:"disabled"`
	NotifyEmail []string `json:"notifyEmail"`
}

func Parse() (Config, error) {
	// ...

	cronConfPath := flag.String("cron-config-file", "cron_config.json", "path of cron config file")
	flag.Parse()

	file, err := os.Open(*cronConfPath)
	if err != nil {
		return Config{}, fmt.Errorf("failed to open config file. %v", err)
	}
	bb, err := ioutil.ReadAll(file)
	if err != nil {
		return Config{}, fmt.Errorf("failed to read config file. %v", err)
	}

	var cc CronConfigs
	if err := json.Unmarshal(bb, &cc); err != nil {
		return Config{}, fmt.Errorf("failed to unmarshal config file. %v", err)
	}

	conf := Config{
		CronConfigs: cc,
	}

	return conf, nil
}
