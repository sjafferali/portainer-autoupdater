package main

import (
	"context"

	"github.com/docker/docker/api/types/swarm"
	"github.com/rs/zerolog"
	"github.com/sjafferali/portainer-autoupdater/internal/portainerapi"
)

func upgradeServices(
	ctx context.Context,
	client portainerapi.Client,
	dryRun bool,
	excludedIDs, includedIDs []string,
	excludedNames, includedNames []string,
	logger zerolog.Logger,
) error {
	endpoints, err := client.Endpoints(ctx, logger)
	if err != nil {
		return err
	}

	for _, endpoint := range endpoints {
		ll := logger.With().
			Int("endpoint_id", int(endpoint.ID)).
			Str("endpoint_name", endpoint.Name).
			Str("endpoint_url", endpoint.URL).
			Logger()

		if len(endpoint.Snapshots) == 0 {
			ll.Trace().Msg("skipping endpoint due to no snapshot")
			continue
		}

		if !endpoint.Snapshots[0].Swarm {
			ll.Trace().Msg("skipping endpoint since not swarm")
			continue
		}

		services, err := client.Services(ctx, int(endpoint.ID), ll)
		if err != nil {
			return err
		}

		if err := processServiceList(
			ctx,
			client,
			services,
			excludedIDs, includedIDs,
			excludedNames, includedNames,
			int(endpoint.ID),
			dryRun,
			ll,
		); err != nil {
			return err
		}
	}
	return err
}

func processServiceList(
	ctx context.Context,
	client portainerapi.Client,
	services []swarm.Service,
	excludedIDs, includedIDs []string,
	excludedNames, includedNames []string,
	endpointID int,
	dryRun bool,
	ll zerolog.Logger,
) error {
	for _, service := range services {
		ll = ll.With().
			Str("service_id", service.ID).
			Str("service_name", service.Spec.Name).
			Logger()

		if includedIDs != nil && !inSlice(includedIDs, service.ID) {
			ll.Trace().Msg("skipping since service ID is not included")
			continue
		}

		if includedNames != nil && !inSlice(includedNames, service.Spec.Name) {
			ll.Trace().Msg("skipping since service name is not included")
			continue
		}

		if excludedIDs != nil && inSlice(excludedIDs, service.ID) {
			ll.Trace().Msg("skipping since service ID is excluded")
			continue
		}

		if excludedNames != nil && inSlice(excludedNames, service.Spec.Name) {
			ll.Trace().Msg("skipping since service ID is excluded")
			continue
		}

		status, err := client.ServiceImageStatus(ctx, service.ID, endpointID, ll)
		if err != nil {
			return err
		}

		ll := ll.With().Str("image_status", status).Logger()

		if status == "updated" {
			ll.Trace().Msg("skipping service since no update needed")
			continue
		}

		ll.Info().Msg("service requires update")

		if dryRun {
			continue
		}

		ll.Info().Msg("updating service")
		if err := client.UpdateService(ctx, service.ID, endpointID, ll); err != nil {
			return err
		}
	}
	return nil
}
