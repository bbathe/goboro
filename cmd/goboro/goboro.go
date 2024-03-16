package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/bbathe/goboro/config"
	"github.com/bbathe/goboro/ui"
)

func main() {
	// show file & location, date & time
	log.SetFlags(log.Ldate | log.Ltime | log.Llongfile)

	// location for log file and default config are in the working directory
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("%+v", err)
	}
	basefn := filepath.Join(wd, "goboro")

	// log to file
	f, err := os.OpenFile(basefn+".log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("%+v", err)
	}
	defer f.Close()
	log.SetOutput(f)

	// process command line
	var configFile string
	flg := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	flg.StringVar(&configFile, "config", "", "Configuration file")
	err = flg.Parse(os.Args[1:])
	if err != nil {
		err := fmt.Errorf("%s\n\nUsage of %s\n  -config string\n    Configuration file", err.Error(), os.Args[0])
		log.Fatalf("%+v", err)
	}

	// read config
	var cfn string
	if len(configFile) > 0 {
		// if user passed a filename, use that
		cfn = configFile
	} else {
		// default config file
		cfn = basefn + ".yaml"
	}

	err = config.Read(cfn)
	if err != nil {
		log.Fatalf("%+v", err)
	}

	// show app, doesn't come back until main window closed
	err = ui.GoBoroWindow()
	if err != nil {
		log.Fatalf("%+v", err)
	}

	// persist config
	err = config.Write()
	if err != nil {
		log.Fatalf("%+v", err)
	}
}
