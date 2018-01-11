package file

import (
	"encoding/xml"
	"io/ioutil"
	"os"
)

const (
	DefaultCartridgeConfigFile = "filedebug_tc_conf.xml"
)

type CartridgeConfig struct {
	XMLName         xml.Name `xml:"filedebug_cartridge_config"`
	DummyIO         bool     `xml:"dummy_id"`
	EmulateReadOnly bool     `xml:"emulate_readonly"`
	Capacity        uint64   `xml:"capacity_mb"`
	CartridgeType   string   `xml:"cart_type"`
	DensityCode     int      `xml:"density_code"`
}

func readCartridgeConfig(path string) (*CartridgeConfig, error) {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cartCfg CartridgeConfig

	if err := xml.Unmarshal(buf, &cartCfg); err != nil {
		return nil, err
	}

	return &cartCfg, nil
}

func writeCartridgeConfig(path string, cfg *CartridgeConfig) error {
	var buf []byte
	buf, err := xml.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n"); err != nil {
		return err
	}

	if _, err := f.Write(buf); err != nil {
		return err
	}

	if _, err := f.WriteString("\n"); err != nil {
		return err
	}

	return nil
}
