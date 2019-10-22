package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"

	"github.com/droundy/goopt"

	"github.com/go-flac/flacpicture"
	"github.com/go-flac/go-flac"
)

var nowPath string
var c chan int
var i int

func main() {
	c = make(chan int, 100)
	nowPath, _ = os.Getwd()
	n := goopt.String([]string{"-n"}, "", "输入")
	o := goopt.String([]string{"-o"}, nowPath, "输出")
	var paths []string

	goopt.Version = "1.0.0"
	goopt.Summary = "flacCover -n 输入文件或目录 -o 输出文件或目录"
	goopt.Parse(nil)

	if *n == "" {
		fmt.Println("输入为空")
		return
	}

	ok, err := pathExists(*n)

	if err != nil {
		log.Fatal(err)
		return
	}

	ok, err = pathExists(*o)

	if err != nil {
		log.Fatal(err)
		return
	}

	if !ok {
		log.Fatal("输入不存在")
		return
	}

	if isFile(*o) {
		log.Fatal("输出需要是个目录")
		return
	}

	if isFile(*n) {
		paths = append(paths, *n)
	} else {
		scan(*n, &paths)
	}

	for _, p := range paths {
		go handle(p, *o)
	}

	var i int

	for {
		<-c

		i++

		fmt.Println(i)

		if i >= len(paths) {
			break
		}
	}

}

func handle(p string, o string) {
	pic, err := extractFLACCover(p)

	if err != nil {
		c <- 1
		log.Println(err)
		return
	}

	if pic == nil {
		log.Println("图是空的")
		c <- 2
		return
	}

	if len(pic.ImageData) <= 0 {
		log.Println("图是空的")
		c <- 2
		return
	}

	f, err := os.Create(path.Join(o, strings.Replace(path.Base(p), ".flac", ".png", -1)))

	defer f.Close()
	if err != nil {
		c <- 2
		log.Println(err)
		return
	}

	ee, err := f.Write(pic.ImageData)

	if err != nil {
		c <- 3
		log.Println(err)
		return
	}

	c <- 4

	fmt.Println(ee)
}

func extractFLACCover(fileName string) (*flacpicture.MetadataBlockPicture, error) {
	f, err := flac.ParseFile(fileName)
	if err != nil {
		return nil, err
	}

	var pic *flacpicture.MetadataBlockPicture
	for _, meta := range f.Meta {
		if meta.Type == flac.Picture {
			pic, err = flacpicture.ParseFromMetaDataBlock(*meta)
			if err != nil {
				return nil, err
			}
		}
	}
	return pic, nil
}

// PathExists 文件是否存在
func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func isFile(f string) bool {
	fi, e := os.Stat(f)
	if e != nil {
		return false
	}
	return !fi.IsDir()
}

func scan(p string, o *[]string) {
	dir, err := ioutil.ReadDir(p)

	if err != nil {
		log.Fatal(err)
		return
	}

	for _, v := range dir {
		if v.IsDir() {
			scan(p+"/"+v.Name(), o)
			continue
		}

		n := strings.LastIndex(v.Name(), "flac")

		if n < 0 {
			continue
		}

		*o = append(*o, path.Clean(p+"/"+v.Name()))
	}
}
