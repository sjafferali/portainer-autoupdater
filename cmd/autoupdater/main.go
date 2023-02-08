package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/sjafferali/portainer-autoupdater/internal/portainerapi"
)

type ConfigSpecification struct {
	Interval      time.Duration `default:"300s" desc:"how often to run app"`
	DryRun        bool          `default:"true" split_words:"true" desc:"only print updates that will be performed"`
	Endpoint      string        `required:"true" desc:"portainer api endpoint"`
	Token         string        `required:"true" desc:"portainer token to use for authentication"`
	LogLevel      string        `default:"INFO" desc:"loglevel to print logs with"`
	EnableStacks  bool          `default:"true" split_words:"true" desc:"enable checking for stack updates"`
	ExcludeStacks []int         `split_words:"true" desc:"stack IDs of stacks that should be excluded from auto update"`
	IncludeStacks []int         `split_words:"true" desc:"stack IDs of stacks that should be included from checks; if not set, all stacks are included"`
}

func main() {
	var s ConfigSpecification
	if err := envconfig.Process("autoupdater", &s); err != nil {
		if err2 := envconfig.Usage("autoupdater", &s); err2 != nil {
			fmt.Println(err2)
		}
		panic(err)
	}

	level, err := zerolog.ParseLevel(strings.ToLower(s.LogLevel))
	if err != nil {
		panic(errors.Wrap(err, "invalid loglevel"))
	}
	zerolog.SetGlobalLevel(level)

	ctx := context.Background()

	client := portainerapi.NewPortainerAPIClient(s.Token, s.Endpoint)
	for {
		if s.EnableStacks {
			if err := upgradeStacks(ctx, client, s.DryRun, s.ExcludeStacks, s.IncludeStacks); err != nil {
				log.Fatal().Err(err).Msg("error running through stacks")
			}
		}

		log.Debug().Dur("interval", s.Interval).Msg("sleeping")
		time.Sleep(s.Interval)
	}
}
