package services

import "context"

type requestContextKey string

const (
	requestContextActorUserID requestContextKey = "request_actor_user_id"
	requestContextIPAddress  requestContextKey = "request_ip_address"
	requestContextUserAgent requestContextKey = "request_user_agent"
)

type RequestContext struct {
	ActorUserID *string
	IPAddress   *string
	UserAgent   *string
}

func WithRequestContext(ctx context.Context, requestContext RequestContext) context.Context {
	if requestContext.ActorUserID != nil {
		ctx = context.WithValue(ctx, requestContextActorUserID, *requestContext.ActorUserID)
	}
	if requestContext.IPAddress != nil {
		ctx = context.WithValue(ctx, requestContextIPAddress, *requestContext.IPAddress)
	}
	if requestContext.UserAgent != nil {
		ctx = context.WithValue(ctx, requestContextUserAgent, *requestContext.UserAgent)
	}
	return ctx
}

func requestContextFromContext(ctx context.Context) RequestContext {
	if ctx == nil {
		return RequestContext{}
	}

	return RequestContext{
		ActorUserID: valueFromContext(ctx, requestContextActorUserID),
		IPAddress:   valueFromContext(ctx, requestContextIPAddress),
		UserAgent:   valueFromContext(ctx, requestContextUserAgent),
	}
}

func valueFromContext(ctx context.Context, key requestContextKey) *string {
	value := ctx.Value(key)
	text, ok := value.(string)
	if !ok || text == "" {
		return nil
	}
	copy := text
	return &copy
}

