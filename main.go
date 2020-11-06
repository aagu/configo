package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/format"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"text/template"
)

//var (
//	pkgInfo *build.Package
//)

var (
	pkgName    = flag.String("pkgName", "config", "set the package name of generated files")
	pkgPath    = flag.String("pkgPath", "./pkg/config", "set the package path to write generated files")
	configFile = flag.String("configFile", "", "set the config file path used to generate files")
)

//利用模板库，生成代码文件
var funcMap = template.FuncMap{
	"DetermineType":  DetermineType,
	"SetConfigValue": SetConfigValue,
	"UppercaseFirst": UppercaseFirst,
}

func main() {
	flag.Parse()
	if len(*configFile) == 0 {
		flag.PrintDefaults()
		log.Fatal("-configFile 必填")
	}
	configs := getConfigs(*configFile)
	delete(configs, "env")
	delete(configs, "profiles")
	configBuff := bytes.NewBufferString("")
	settingBuff := bytes.NewBufferString("")
	genHeader(configBuff)
	genHeader(settingBuff)
	genExportVars(settingBuff, configs)
	addedStructs := inspectStructs(configs)
	genStructs(configBuff, map[interface{}]interface{}{
		"envConfig": configs,
	}, true)
	genStructs(configBuff, addedStructs, false)

	configSrc, err := format.Source(configBuff.Bytes())
	if err != nil {
		log.Fatalf("error formating source code, %v", err)
	}
	settingSrc, err := format.Source(settingBuff.Bytes())
	if err != nil {
		log.Fatalf("error formating source code, %v", err)
	}

	//保存到文件
	err = os.MkdirAll(*pkgPath, os.ModePerm)
	if err != nil {
		log.Fatalf("writing output: %s", err)
	}
	configName := filepath.Join(*pkgPath, "config.go")
	settingName := filepath.Join(*pkgPath, "settings.go")
	err = ioutil.WriteFile(configName, configSrc, 0644)
	if err != nil {
		log.Fatalf("writing output: %s", err)
	}
	err = ioutil.WriteFile(settingName, settingSrc, 0644)
	if err != nil {
		log.Fatalf("writing output: %s", err)
	}
}

func inspectStructs(in map[interface{}]interface{}) map[interface{}]interface{} {
	addedStructs := make(map[interface{}]interface{})
	for key, val := range in {
		if reflect.TypeOf(val).Kind() == reflect.Map {
			addedStructs[key] = val
			structs := inspectStructs(val.(map[interface{}]interface{}))
			for k, v := range structs {
				addedStructs[k] = v
			}
		}
	}
	return addedStructs
}

func getConfigs(filePath string) map[interface{}]interface{} {
	_, err := os.Stat(filePath)
	if err != nil {
		log.Fatalf("config file path %s not exist!", filePath)
	}
	fileBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatalf("reading config file failed, %v", err)
	}
	var configMap map[interface{}]interface{}
	_ = yaml.Unmarshal(fileBytes, &configMap)
	return configMap
}

func DetermineType(v interface{}) string {
	typ := reflect.TypeOf(v).Kind()
	switch typ {
	case reflect.Int:
		return "int"
	case reflect.Bool:
		return "bool"
	case reflect.Float64:
		return "float64"
	case reflect.Map:
		return ""
	default:
		return "string"
	}
}

func SetConfigValue(typ string, value interface{}) string {
	switch typ {
	case "string":
		return fmt.Sprintf("\"%s\"", value)
	default:
		return fmt.Sprintf("%v", value)
	}
}

func UppercaseFirst(v string) string {
	return fmt.Sprintf("%s%s", strings.ToUpper(v[:1]), v[1:])
}

func genExportVars(buff *bytes.Buffer, configs map[interface{}]interface{}) {
	primaryTypesTpl, err := template.New("vars.tmpl").Funcs(funcMap).ParseFiles("./templates/vars.tmpl")
	if err != nil {
		log.Fatal(err)
	}
	err = primaryTypesTpl.Execute(buff, configs)
	if err != nil {
		log.Fatal(err)
	}
}

func genHeader(buff *bytes.Buffer) {
	primaryTypesTpl, err := template.New("header.tmpl").ParseFiles("./templates/header.tmpl")
	if err != nil {
		log.Fatal(err)
	}
	err = primaryTypesTpl.Execute(buff, pkgName)
	if err != nil {
		log.Fatal(err)
	}
}

func genStructs(buff *bytes.Buffer, types map[interface{}]interface{}, exported bool) {
	data := map[string]interface{}{
		"types":    types,
		"exported": exported,
	}
	//利用模板库，生成代码文件
	primaryTypesTpl, err := template.New("struct.tmpl").Funcs(funcMap).ParseFiles("./templates/struct.tmpl")
	//primaryTypesTpl, err := template.New("").Funcs(funcMap).Parse(strTmp)
	if err != nil {
		log.Fatal(err)
	}
	//primaryTypesTpl.Funcs(funcMap)
	err = primaryTypesTpl.Execute(buff, data)
	if err != nil {
		log.Fatal(err)
	}
}
