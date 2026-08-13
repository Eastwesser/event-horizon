package interceptor

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const MDRole = "x-user-role"

// RequireRoles rejects RPCs whose method suffix matches protectedSuffixes unless
// metadata x-user-role is one of allowed.
func RequireRoles(allowed []string, protectedSuffixes []string) grpc.UnaryServerInterceptor {
	allow := map[string]struct{}{}
	for _, r := range allowed {
		allow[strings.ToLower(r)] = struct{}{}
	}
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		need := false
		for _, s := range protectedSuffixes {
			if strings.HasSuffix(info.FullMethod, s) {
				need = true
				break
			}
		}
		if !need {
			return handler(ctx, req)
		}
		md, _ := metadata.FromIncomingContext(ctx)
		role := ""
		if vals := md.Get(MDRole); len(vals) > 0 {
			role = strings.ToLower(vals[0])
		}
		if _, ok := allow[role]; !ok {
			return nil, status.Error(codes.PermissionDenied, "insufficient permissions")
		}
		return handler(ctx, req)
	}
}
