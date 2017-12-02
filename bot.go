package main

import (
	"crypto/tls"
	"fmt"
	"regexp"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/asdine/storm"
	"github.com/thoj/go-ircevent"
)

type Config struct {
	nick string
	user string
	name string

	addr   *Addr
	debug  bool
	logdir string
}

type Bot struct {
	config *Config
	conn   *irc.Connection

	network string
	loggers map[string]map[string]Logger
}

func NewBot(config *Config) *Bot {
	return &Bot{
		config:  config,
		loggers: make(map[string]map[string]Logger),
	}
}

func (b *Bot) onInvite(e *irc.Event) {
	var channel Channel
	err := db.One("Name", e.Arguments[0], &channel)
	if err != nil && err == storm.ErrNotFound {
		channel = NewChannel(e.Arguments[1])
		err := db.Save(&channel)
		if err != nil {
			log.Fatalf("error saving channel to db: %s", err)
		}
	} else if err != nil {
		log.Fatalf("error looking up channel in db: %s", err)
	}

	log.Infof("Requested to join %s", channel.Name)

	b.conn.Join(channel.Name)
	b.conn.Mode(channel.Name)
}

func (b *Bot) onConnected(e *irc.Event) {
	var channels []Channel

	log.Info("Connected!")

	p := regexp.MustCompile("^[Ww]elcome to the (.*) Internet Relay Network")
	matches := p.FindStringSubmatch(e.Message())
	if len(matches) == 2 {
		b.network = matches[1]
	}

	err := db.All(&channels)
	if err != nil {
		log.Fatalf("error loading channels from db: %s", err)
	}

	for _, channel := range channels {
		b.conn.Join(channel.Name)
		b.conn.Mode(channel.Name)
		log.Infof("Joined %s", channel.Name)
	}
}

func (b *Bot) onMessage(e *irc.Event) {
	if e.Arguments[0][0] != '#' {
		return
	}

	ts := time.Now().Format("15:04:05")
	channel := e.Arguments[0]
	logger, ok := b.loggers[b.network][channel]
	if !ok {
		logger, err := NewFileLogger(b.config.logdir, b.network, channel)
		if err != nil {
			log.Errorf(
				"failed to create logger for %s %s: %s",
				b.network, channel, err,
			)
			return
		}
		network := b.loggers[b.network]
		if network == nil {
			b.loggers[b.network] = make(map[string]Logger)
		}
		b.loggers[b.network][channel] = logger
		logger.Log(fmt.Sprintf("[%s] <%s> %s\n", ts, e.User, e.Message()))
	} else {
		logger.Log(fmt.Sprintf("[%s] <%s> %s\n", ts, e.User, e.Message()))
	}
}

func (b *Bot) onJoin(e *irc.Event) {
	log.Debugf("onJoin: %v", e)
}

func (b *Bot) onPart(e *irc.Event) {
	log.Debugf("onPart: %v", e)
}

func (b *Bot) onQuit(e *irc.Event) {
	log.Debugf("onQuit: %v", e)
}

func (b *Bot) Run() error {
	b.conn = irc.IRC(b.config.nick, b.config.user)
	b.conn.RealName = b.config.name

	b.conn.VerboseCallbackHandler = b.config.debug
	b.conn.Debug = b.config.debug

	b.setupCallbacks()

	b.conn.UseTLS = b.config.addr.UseTLS
	b.conn.KeepAlive = 30 * time.Second
	b.conn.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	err := b.conn.Connect(b.config.addr.String())
	if err != nil {
		return err
	}

	b.conn.Loop()
	return nil
}

func (b *Bot) setupCallbacks() {
	b.conn.AddCallback("001", b.onConnected)
	b.conn.AddCallback("INVITE", b.onInvite)
	b.conn.AddCallback("JOIN", b.onJoin)
	b.conn.AddCallback("PART", b.onPart)
	b.conn.AddCallback("QUIT", b.onQuit)
	b.conn.AddCallback("PRIVMSG", b.onMessage)
}
