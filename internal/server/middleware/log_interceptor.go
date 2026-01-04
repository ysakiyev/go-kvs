package middleware

import (
	"context"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

// UnaryServerLoggingInterceptor logs information about unary RPCs.
func UnaryServerLoggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	startTime := time.Now()

	lastIdx := strings.LastIndex(info.FullMethod, "/")
	cmdStr := info.FullMethod[lastIdx+1:]

	// Call the handler to process the unary RPC.
	resp, err := handler(ctx, req)
	if err != nil {
		log.Error().Msgf("%v, Cmd: %s, took %s", err, cmdStr, time.Since(startTime))
		return resp, err
	}

	// Log information about the unary RPC.
	log.Info().Msgf("Cmd %s took %s", cmdStr, time.Since(startTime))
	return resp, err
}
