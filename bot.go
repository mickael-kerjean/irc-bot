package cadmus

import (
	"crypto/tls"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/asdine/storm"
	"github.com/robfig/cron"
	"github.com/thoj/go-ircevent"
)

type Config struct {
	Nick     string
	User     string
	Name     string
	Password string
	Chan     []string
	Debug    bool
	DBPath   string
	LogPath  string
}

type Bot struct {
	addr    *Addr
	config  *Config
	db      *storm.DB
	cron    *cron.Cron
	conn    *irc.Connection
	network string
	loggers *ChannelLoggerMap
}

func NewBot(addr string, config *Config) *Bot {
	parsedAddr, err := ParseAddr(addr)
	if err != nil {
		log.Fatalf("error parsing addr %s: %s", addr, err)
	}

	return &Bot{
		addr:    parsedAddr,
		config:  config,
		loggers: NewChannelLoggerMap(),
	}
}

func (b *Bot) getLogger(channel string) (logger Logger, err error) {
	logger = b.loggers.Get(channel)
	if logger == nil {
		logger, err = NewFileLogger(b.config.LogPath, b.network, channel)
		if err != nil {
			return nil, err
		}
		b.loggers.Add(logger)
	}
	return
}

func (b *Bot) onInvite(e *irc.Event) {
	var channel Channel
	err := b.db.One("Name", e.Arguments[0], &channel)
	if err != nil && err == storm.ErrNotFound {
		channel = NewChannel(e.Arguments[1])
		err := b.db.Save(&channel)
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
	log.Info("Connected!")
	for _, channel := range b.config.Chan {
		b.conn.Join(channel)
		b.conn.Mode(channel)
		log.Infof("Joined %s", channel)
	}
}

func (b *Bot) onMessage(e *irc.Event) {
	if e.Arguments[0][0] != '#' {
		return
	}

	channel := e.Arguments[0]

	logger, err := b.getLogger(channel)
	if err != nil {
		log.Errorf(
			"error getting logger for %s on %s: %s",
			channel, b.network, err,
		)
		return
	}

	logger.LogMessage(e.User, e.Message())
}

func (b *Bot) Run() error {
	db, err := storm.Open(b.config.DBPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	b.db = db

	b.cron = cron.New()
	b.cron.AddFunc("@daily", func() {
		b.loggers.Range(func(channel string, logger Logger) bool {
			logger.Rotate()
			log.Infof(
				"Logger rotated for %s on %s",
				logger.Channel(), logger.Network(),
			)
			return true
		})
	})
	b.cron.Start()

	b.conn = irc.IRC(b.config.Nick, b.config.User)
	b.conn.RealName = b.config.Name
	if b.config.Password != "" {
		b.conn.Password = b.config.Password
	}

	b.conn.VerboseCallbackHandler = b.config.Debug
	b.conn.Debug = b.config.Debug

	b.setupCallbacks()

	b.conn.UseTLS = b.addr.UseTLS
	b.conn.KeepAlive = 30 * time.Second
	b.conn.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	err = b.conn.Connect(b.addr.String())
	if err != nil {
		return err
	}

	b.conn.Loop()
	return nil
}

func (b *Bot) setupCallbacks() {
	b.conn.AddCallback("001", b.onConnected)
	b.conn.AddCallback("INVITE", b.onInvite)
	b.conn.AddCallback("PRIVMSG", b.onMessage)
}
