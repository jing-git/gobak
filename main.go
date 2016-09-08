package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"qiniupkg.com/api.v7/conf"
	"qiniupkg.com/api.v7/kodo"
	"qiniupkg.com/api.v7/kodocli"
	"strings"
	"time"
)

const defCfgFile = "gobak.cfg"

type UplInfo struct {
	BakFile string
	AK      string
	SK      string
	Bucket  string
	Key     string
}

type PutRet struct {
	Hash string `json:"hash"`
	Key  string `json:"key"`
}

func getCurDir() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return ""
	}
	return dir + "/"
}

func do7zBak(cfgFile string, ui *UplInfo) bool {
	bytes, _ := ioutil.ReadFile(cfgFile)
	if bytes != nil {
		lines := strings.Split(string(bytes), "\n")
		max := len(lines)
		/*
			AK
			SK
			Bucket
			file1
			file2
			..
		*/
		from := 3
		if max > from {
			ui.AK = lines[0]
			ui.SK = lines[1]
			ui.Bucket = lines[2]
			//bak 3 days' files
			ui.Key = fmt.Sprintf("%d", time.Now().Weekday()%3) + ".7z"
			ui.BakFile = getCurDir() + ui.Key
			os.Remove(ui.BakFile)
			for i := from; i < max; i++ {
				if lines[i] == "" || lines[i] == "\n" {
					continue
				}
				cmd := exec.Command("7z", "a", "-t7z", "-mhe=on", ui.BakFile, lines[i])
				if _, err := cmd.Output(); err != nil {
					fmt.Println("7z fail:", lines[i], err)
				}
			}

			return true
		}
	}

	return false
}

func QNUpl(ui *UplInfo) bool {
	conf.ACCESS_KEY = ui.AK
	conf.SECRET_KEY = ui.SK

	c := kodo.New(0, nil)
	policy := &kodo.PutPolicy{
		Scope:   ui.Bucket + ":" + ui.Key,
		Expires: 3600,
	}
	token := c.MakeUptoken(policy)

	zone := 0
	uploader := kodocli.NewUploader(zone, nil)

	var ret PutRet
	filepath := ui.BakFile
	res := uploader.PutFile(nil, &ret, token, ui.Key, filepath, nil)
	if res != nil {
		fmt.Println("upl fail:", res)
		return false
	}

	return true
}

func main() {
	var ui UplInfo

	fmt.Println("7z..")
	if do7zBak(getCurDir()+defCfgFile, &ui) {
		fmt.Println("upl..")
		if QNUpl(&ui) {
			fmt.Println("baked.")
		}
		os.Remove(ui.BakFile)
	}
}
