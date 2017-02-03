package main

import (
	"flag"
	"fmt"
	L "github.com/absinsekt/mobile-icon-cropper/lib"
	"gopkg.in/gographics/imagick.v2/imagick"
	"os"
	"path/filepath"
)

func main() {
	var (
		srcImageFile    string
		configFile      string
		targetDirectory string
		concurrency     uint
	)

	flag.StringVar(&configFile, "f", "config.yaml", "configuration file in yaml format (default: config.yaml)")
	flag.StringVar(&targetDirectory, "d", "out", "target directory (default: out)")
	flag.UintVar(&concurrency, "c", 5, "batch concurrency (default: 5)")
	flag.Parse()

	if concurrency > 10 || concurrency < 1 {
		concurrency = 5
	}

	args := flag.Args()

	if len(args) == 0 {
		notifyError("source file not set")
	} else {
		srcImageFile = args[0]
	}

	conf := L.ConfigProvider{}
	conf.Initialize(configFile)

	imagick.Initialize()
	defer imagick.Terminate()

	mw := imagick.NewMagickWand()

	if err := mw.ReadImage(srcImageFile); err != nil {
		notifyError(err)
	}

	if err := os.MkdirAll(targetDirectory, 0755); err != nil {
		notifyError(err)
	}

	batch := make(chan L.MagickCropper, concurrency)
	done := make(chan string)

	for _, set := range conf.ConfigData.Sets {
		for _, icon := range set.Icons {
			go func(tf string, icfg L.IconConfig) {
				crp := <-batch

				crp.SmartCrop(icfg.Width, icfg.Height)
				crp.ShapeImage(icfg.Type, 10)

				if err := crp.MagickWand.WriteImage(filepath.Join(targetDirectory, tf)); err != nil {
					notifyError(err)
				}

				done <- tf

			}(fmt.Sprintf("%s_%dx%d.png", set.Prefix, icon.Width, icon.Height), icon)

			batch <- L.MagickCropper{mw.Clone()}
		}
	}

	for i := 0; i < conf.ConfigData.Length(); i++ {
		fmt.Printf("file ready: %s\n", <-done)
	}
}

func notifyError(err interface{}) {
	panic(err)
}
