package slgo

import "github.com/bclswl0827/slgo/handlers"

func New(provider handlers.SeedLinkProvider, consumer handlers.SeedLinkConsumer, hooks handlers.SeedLinkHooks) SeedLinkServer {
	return SeedLinkServer{
		hooks:    hooks,
		Provider: provider,
		Consumer: consumer,
	}
}
