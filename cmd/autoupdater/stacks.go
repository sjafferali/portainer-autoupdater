package main

import (
	"context"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/sjafferali/portainer-autoupdater/internal/portainerapi"
)

func upgradeStacks(
	ctx context.Context,
	client portainerapi.Client,
	dryRun bool,
	excludedIDs, includedIDs []int,
	excludedNames, includedNames []string,
	logger zerolog.Logger,
) error {
	stacks, err := client.Stacks(ctx, logger)
	if err != nil {
		return errors.Wrap(err, "error getting stacks")
	}

	for _, i := range stacks {
		stackID := int(i.ID)

		ll := logger.With().
			Str("name", i.Name).
			Int("stack_id", stackID).
			Logger()

		if includedIDs != nil && !inSlice(includedIDs, stackID) {
			ll.Trace().Msg("skipped since stack ID is not included")
			continue
		}

		if includedNames != nil && !inSlice(includedNames, i.Name) {
			ll.Trace().Msg("skipped since stack name is not included")
			continue
		}

		if excludedIDs != nil && inSlice(excludedIDs, stackID) {
			ll.Trace().Msg("skipped since stack id is excluded")
			continue
		}

		if excludedNames != nil && inSlice(excludedNames, i.Name) {
			ll.Trace().Msg("skipped since stack name is excluded")
			continue
		}

		status, err := client.StackImageStatus(ctx, stackID, ll)
		if err != nil {
			ll.Error().Err(err).Msg("error getting image status")
			continue
		}
		ll = ll.With().Str("status", status).Logger()

		if status != "outdated" {
			ll.Debug().Msg("no update needed")
			continue
		}
		ll.Info().Msg("stack needs update")
		if !dryRun {
			ll.Info().Msg("updating")
			if err := client.UpdateStack(ctx, stackID, ll); err != nil {
				ll.Error().Err(err).Msg("error updating stack")
				continue
			}
		}
	}
	return nil
}
