package main

import (
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/ini.v1"
)

var cfg *Config = nil

func GetConfig() *Config {
	if cfg == nil {
		InitConfig()
	}
	return cfg
}

type Config struct {
	RemoveString []string
	InputDir     string
	OutputDir    string
}

func InitConfig() error {
	ctype := os.Getenv("CONFIG_TYPE")

	if cfg == nil {
		cfg = new(Config)
	}

	var err error
	if ctype == "env" {
		//	err = cfg.initEnvConfig()
	} else {
		err = cfg.initConf()
	}

	return err
}

func (conf *Config) initConf() error {
	exePath, err := os.Executable()
	if err != nil {
		return err
	}

	exeDir := filepath.Dir(exePath)

	cfg, err := ini.Load(exeDir + "/config.ini")
	if err != nil {
		return err
	}

	tmp := cfg.Section("prompt").Key("removeString").String()
	tmpArr := strings.Split(tmp, ",")
	for _, v := range tmpArr {
		conf.RemoveString = append(conf.RemoveString, strings.TrimSpace(v))
	}

	conf.InputDir = cfg.Section("prompt").Key("inputDir").String()
	conf.OutputDir = cfg.Section("prompt").Key("outputDir").String()

	return nil
}
