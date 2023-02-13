package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Conf struct {
	ID     string
	Secret string
}

func confPath() (string, error) {
	mDir, err := mujifyDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(mDir, "config.json"), nil
}

func newConf() (*Conf, error) {
	cPath, err := confPath()
	if err != nil {
		return nil, err
	}
	var c *Conf
	if _, err := os.Stat(cPath); err != nil {
		c, err = initConf()
		if err != nil {
			return nil, err
		}
	} else {
		cFile, err := os.Open(cPath)
		if err != nil {
			return nil, err
		}
		defer cFile.Close()
		json.NewDecoder(cFile).Decode(c)
	}
	return c, nil
}

func initConf() (*Conf, error) {
	cPath, err := confPath()
	if err != nil {
		return nil, err
	}
	c := new(Conf)
	fmt.Println(cPath, "not found...")
	fmt.Println("please input your app's ID and SECRET")
	fmt.Print("ID: ")
	fmt.Scan(&c.ID)
	fmt.Print("SECRET: ")
	fmt.Scan(&c.Secret)

	cFile, err := os.Create(cPath)
	if err != nil {
		return nil, err
	}
	defer cFile.Close()
	json.NewEncoder(cFile).Encode(c)

	return c, nil
}
