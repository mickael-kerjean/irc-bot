package main

import (
	"fmt"
	"github.com/namsral/flag"
	"github.com/mickael-kerjean/prologic_cadmus_fork"
	log "github.com/sirupsen/logrus"
	"os"
	"strings"
)

const (
	nick = "filestash"
	user = "filestash"
	name = "Filestash, Bot"
)

func main() {
	var (
		version bool
		config  string
		debug   bool
		dbpath  string
		logpath string
	)

	flag.BoolVar(&version, "v", false, "display version information")
	flag.StringVar(&config, "c", "", "config file")
	flag.BoolVar(&debug, "d", false, "debug logging")
	flag.StringVar(&dbpath, "dbpath", "/app/cadmus.db", "path to database")
	flag.StringVar(&logpath, "logpath", "/app/logs", "path to store logs")
	flag.Parse()

	if debug {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	if version {
		fmt.Printf(cadmus.FullVersion())
		os.Exit(0)
	}

	if flag.NArg() < 1 {
		log.Fatalf("Ussage: %s <address>[:port]", os.Args[0])
	}

	bot := cadmus.NewBot(flag.Arg(0), &cadmus.Config{
		Chan:     strings.Split(os.Getenv("BOT_CHANNELS"), ","),
		Nick:     os.Getenv("BOT_NICK"),
		User:     os.Getenv("BOT_USER"),
		Name:     os.Getenv("BOT_REALNAME"),
		Password: os.Getenv("BOT_PASSWORD"),
		Debug:    debug,
		DBPath:   dbpath,
		LogPath:  logpath,
	})
	log.Fatal(bot.Run())
}
