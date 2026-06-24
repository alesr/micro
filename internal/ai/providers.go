package ai

import "context"

type Request struct {
	BeforeCursor string
	AfterCursor  string
	FileType     string
	FileName     string
}

type Response struct {
	Text string
}

type Provider interface {
	Complete(ctx context.Context, req Request) (*Response, error)
	Name() string
}
