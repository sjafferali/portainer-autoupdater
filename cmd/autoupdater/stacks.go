package main

import (
	"context"
	"fmt"
	"github.com/grab/async"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/sjafferali/portainer-autoupdater/internal/portainerapi"
	"time"
)

func upgradeStacks(
	ctx context.Context,
	client portainerapi.Client,
	dryRun bool,
	excludedIDs, includedIDs []int,
	excludedNames, includedNames []string,
	interval time.Duration,
	logger zerolog.Logger,
) error {

	stacks, err := client.Stacks(ctx, logger)
	if err != nil {
		return errors.Wrap(err, "error getting stacks")
	}
	logger.Info().Int("stacks_count", len(stacks)).Msg("found stacks")

	tasks := make([]async.Task, 0)
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

		task := getTaskForStack(ctx, client, dryRun, stackID, ll)
		tasks = append(tasks, task)
	}

	logger.Info().Int("stacks_to_check", len(tasks)).Msg("stacks to check")
	tasktracker := async.Spread(ctx, interval, tasks)
	_, _ = tasktracker.Outcome() // Wait

	// Make sure all tasks are done
	for _, singletask := range tasks {
		v, _ := singletask.Outcome()
		fmt.Println(v)
	}
	return nil
}

func getTaskForStack(
	ctx context.Context,
	client portainerapi.Client,
	dryRun bool,
	stackID int,
	ll zerolog.Logger,
) async.Task {
	task := async.NewTask(func(context.Context) (interface{}, error) {
		updated := false
		ll.Trace().Msg("checking stack")
		status, err := client.StackImageStatus(ctx, stackID, ll)
		if err != nil {
			ll.Error().Err(err).Msg("error getting image status")
			return nil, err
		}
		ll = ll.With().Str("status", status).Logger()

		if status != "outdated" {
			ll.Debug().Msg("no update needed")
			return nil, err
		}
		ll.Info().Msg("stack needs update")
		if !dryRun {
			ll.Info().Msg("updating")
			if err := client.UpdateStack(ctx, stackID, ll); err != nil {
				ll.Error().Err(err).Msg("error updating stack")
				return nil, err
			}
			updated = true
		}

		msg := fmt.Sprintf("stack %d not updated", stackID)
		if updated {
			msg = fmt.Sprintf("stack %d updated", stackID)
		}
		return msg, nil
	})
	return task
}
