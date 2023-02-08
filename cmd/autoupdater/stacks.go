package main

import (
	"context"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/sjafferali/portainer-autoupdater/internal/portainerapi"
)

func inSlice(in []int, value int) bool {
	for _, i := range in {
		if value == i {
			return true
		}
	}
	return false
}

func upgradeStacks(ctx context.Context, client portainerapi.Client, dryRun bool, excluded, included []int) error {
	stacks, err := client.Stacks(ctx)
	if err != nil {
		return errors.Wrap(err, "error getting stacks")
	}

	for _, i := range stacks {
		stackID := int(i.ID)

		ll := log.With().
			Str("name", i.Name).
			Int("stack_id", stackID).
			Logger()

		if included != nil && !inSlice(included, stackID) {
			ll.Trace().Msg("skipped since INCLUDE_STACKS is set and stack id is missing")
			continue
		}

		if excluded != nil && inSlice(excluded, stackID) {
			ll.Trace().Msg("skipped since stack id excluded by EXCLUDE_STACKS")
			continue
		}

		status, err := client.StackImageStatus(ctx, stackID)
		if err != nil {
			ll.Error().Err(err).Msg("error getting image status")
			continue
		}
		ll = ll.With().Str("status", status).Logger()

		if status != "outdated" {
			ll.Debug().Msg("no update needed")
			continue
		}

		ll.Info().Msg("image needs update")
		if !dryRun {
			ll.Info().Msg("updating")
			if err := client.UpdateStack(ctx, stackID); err != nil {
				ll.Error().Err(err).Msg("error updating stack")
				continue
			}
		}
	}
	return nil
}
