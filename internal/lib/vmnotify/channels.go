package vmnotify

import "context"

type Channel string

type Event struct {}

type Starter interface {
	Start(ctx context.Context, events <-chan Event) Channel
}