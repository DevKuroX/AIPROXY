package router

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/DevKuroX/AIPROXY/internal/handlers/chat"
	"github.com/DevKuroX/AIPROXY/internal/stream"
)

type ChatRouter struct {
	streamingHandler    *chat.StreamingHandler
	nonStreamingHandler *chat.NonStreamingHandler
	sseToJsonConverter  *chat.SSEToJSONConverter
	errorHandler        *chat.ErrorHandler
}

func NewChatRouter(
	streamingHandler *chat.StreamingHandler,
	nonStreamingHandler *chat.NonStreamingHandler,
	sseToJsonConverter *chat.SSEToJSONConverter,
	errorHandler *chat.ErrorHandler,
) *ChatRouter {
	return &ChatRouter{
		streamingHandler:    streamingHandler,
		nonStreamingHandler: nonStreamingHandler,
		sseToJsonConverter:  sseToJsonConverter,
		errorHandler:        errorHandler,
	}
}

func (r *ChatRouter) RouteChatRequest(
	ctx context.Context,
	w http.ResponseWriter,
	req *http.Request,
	upstreamResp *http.Response,
	config *stream.StreamConfig,
) error {
	var chatReq chat.ChatRequest
	if err := json.NewDecoder(req.Body).Decode(&chatReq); err != nil {
		r.errorHandler.HandleError(ctx, w, err, config.Stream)
		return err
	}

	if chatReq.Stream || config.Stream {
		return r.streamingHandler.HandleStreaming(ctx, w, upstreamResp, config)
	}

	return r.nonStreamingHandler.HandleNonStreaming(ctx, w, upstreamResp, config)
}

func (r *ChatRouter) ConvertSSEToJSON(
	ctx context.Context,
	upstreamResp *http.Response,
	format stream.StreamFormat,
) ([]byte, *stream.Usage, error) {
	return r.sseToJsonConverter.Convert(ctx, upstreamResp.Body, format)
}

func (r *ChatRouter) HandleChatError(
	ctx context.Context,
	w http.ResponseWriter,
	err error,
	streaming bool,
) {
	r.errorHandler.HandleError(ctx, w, err, streaming)
}
