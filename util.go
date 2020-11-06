package main

import (
	"github.com/imdario/mergo"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"reflect"
)

func readFile(filePath string) ([]byte, error) {
	_, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}
	return ioutil.ReadFile(filePath)
}

func Merge(filePath string, config interface{}) {
	bytes, err := readFile(filePath)
	if err != nil {
		log.Fatalf("read config file failed, %v", err)
	}
	mappedYaml := make(map[interface{}]interface{})
	err = yaml.Unmarshal(bytes, &mappedYaml)
	if err != nil {
		log.Fatalf("dump yaml failed, %v", err)
	}
	err = yaml.Unmarshal(bytes, config)
	if err != nil {
		log.Fatalf("dump yaml to struct failed, %v", err)
	}
	profileMaps, ok := mappedYaml["profiles"]
	if !ok {
		return
	}
	profileDecls, ok := profileMaps.(map[interface{}]interface{})
	if !ok {
		log.Printf("unknown type of profile declaration, expect map, got %v", reflect.TypeOf(profileMaps).Kind())
	}
	marshal, err := yaml.Marshal(profileDecls[mappedYaml["env"]])
	if err != nil {
		log.Fatalf("marshal failed, %v", err)
	}
	configTyp := reflect.TypeOf(config)
	if configTyp.Kind() == reflect.Ptr {
		configTyp = configTyp.Elem()
	}
	override := reflect.New(reflect.TypeOf(config).Elem()).Interface()
	//var override *envConfig
	err = yaml.Unmarshal(marshal, override)
	if err != nil {
		log.Fatalf("dump profile to struct failed, %v", err)
	}
	err = mergo.Merge(config, override, mergo.WithOverride)
	if err != nil {
		log.Fatalf("failed to merge config, %v", err)
	}
}
