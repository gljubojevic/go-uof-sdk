package sdk

import (
	"context"
	"time"

	"github.com/minus5/go-uof-sdk"
	"github.com/minus5/go-uof-sdk/api"
	"github.com/minus5/go-uof-sdk/pipe"
	"github.com/minus5/go-uof-sdk/queue"
)

var defaultLanuages = uof.Languages("en,de")

type Config struct {
	BookmakerID string
	Token       string
	Staging     bool
	Languages   []uof.Lang
	Fixtures    time.Time
	Recovery    []uof.ProducerChange
	Stages      []pipe.StageHandler
	Replay      func(*api.ReplayApi) error
}

// Option sets attributes on the Config.
type Option func(*Config)

// Run starts uof connector.
//
// Call to Run blocks until stopped by context, or error occured.
// Order in wich options are set is not important.
// Credentials and one of Callback or Pipe are functional minimum.
func Run(ctx context.Context, options ...Option) error {
	c := config(options...)
	qc, apiConn, err := connect(ctx, c)
	if err != nil {
		return err
	}
	if c.Replay != nil {
		rpl, err := api.Replay(ctx, c.Token)
		if err != nil {
			return err
		}
		if err := c.Replay(rpl); err != nil {
			return err
		}
	}

	stages := []pipe.StageHandler{
		pipe.Markets(apiConn, c.Languages),
		pipe.Fixture(apiConn, c.Languages, c.Fixtures),
		pipe.Player(apiConn, c.Languages),
		pipe.BetStop(),
	}
	if len(c.Recovery) > 0 {
		stages = append(stages, pipe.Recovery(apiConn, c.Recovery))
	}
	stages = append(stages, c.Stages...)

	errc := pipe.Build(
		queue.WithReconnect(ctx, qc),
		stages...,
	)
	return firstErr(errc)
}

func firstErr(errc <-chan error) error {
	var err error
	for e := range errc {
		if err == nil {
			err = e
		}
	}
	return err
}

func config(options ...Option) Config {
	c := &Config{}
	for _, o := range options {
		o(c)
	}
	// defaults
	if len(c.Languages) == 0 {
		c.Languages = defaultLanuages
	}
	return *c
}

// TODO: pojednostavi ovo
func connect(ctx context.Context, c Config) (*queue.Connection, *api.Api, error) {
	if c.Replay != nil {
		conn, err := queue.DialReplay(ctx, c.BookmakerID, c.Token)
		if err != nil {
			return nil, nil, err
		}
		stg, err := api.Staging(ctx, c.Token)
		if err != nil {
			return nil, nil, err
		}
		return conn, stg, nil
	}

	if c.Staging {
		conn, err := queue.DialStaging(ctx, c.BookmakerID, c.Token)
		if err != nil {
			return nil, nil, err
		}
		stg, err := api.Staging(ctx, c.Token)
		if err != nil {
			return nil, nil, err
		}
		return conn, stg, nil
	}

	conn, err := queue.Dial(ctx, c.BookmakerID, c.Token)
	if err != nil {
		return nil, nil, err
	}
	stg, err := api.Production(ctx, c.Token)
	if err != nil {
		return nil, nil, err
	}
	return conn, stg, nil
}

// Credentials for establishing connection to the uof queue and api.
func Credentials(bookmakerID, token string) Option {
	return func(c *Config) {
		c.BookmakerID = bookmakerID
		c.Token = token
	}
}

// Languages for api calls.
//
// Statefull messages (markets, players, fixtures) will be served in all this
// languages. Each language requires separate call to api. If not specified
// `defaultLanguages` will be used.
func Languages(langs []uof.Lang) Option {
	return func(c *Config) {
		c.Languages = langs
	}
}

// Staging forces use of staging environment instead of production.
func Staging() Option {
	return func(c *Config) {
		c.Staging = true
	}
}

// Replay forces use of replay environment.
// Callback will be called to start replay after establishing connection.
func Replay(cb func(*api.ReplayApi) error) Option {
	return func(c *Config) {
		c.Replay = cb
	}
}

// Pipe sets chan handler for all messages.
// Can be called multiple times.
func Pipe(s pipe.StageHandler) Option {
	return func(c *Config) {
		c.Stages = append(c.Stages, s)
	}
}

// Callback sets handler for all messages.
//
// If returns error will break the pipe and force exit from sdk.Run.
// Can be called multiple times.
func Callback(cb func(m *uof.Message) error) Option {
	return func(c *Config) {
		c.Stages = append(c.Stages, pipe.Simple(cb))
	}
}

// Recovery starts recovery for each producer
//
// It is responsibility of SDK consumer to track the last timestamp of the
// successfuly consumed message for each producer. On startup this timestamp is
// sent here and SDK will request recovery; get all the messages after that ts.
//
// Ref: https://docs.betradar.com/display/BD/UOF+-+Recovery+using+API
func Recovery(pc []uof.ProducerChange) Option {
	return func(c *Config) {
		c.Recovery = pc
	}
}

// Fixtures gets pre-match fixtures at start-up.
//
// It gets fixture for all matches which starts before `to` time.
// There is a special endpoint to get almost all fixtures before initiating
// recovery. This endpoint is designed to significantly reduce the number of API
// calls required during recovery.
//
// Ref: https://docs.betradar.com/display/BD/UOF+-+Fixtures+in+the+API
func Fixtures(to time.Time) Option {
	return func(c *Config) {
		c.Fixtures = to
	}
}