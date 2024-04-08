package jwt

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/form3tech-oss/jwt-go"
	"github.com/golang/glog"
	"github.com/lainio/err2"
	"github.com/lainio/err2/try"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Define a secure key string used
// as a salt when hashing our tokens.
// Please make your own way more secure than this,
// use a randomly generated md5 hash or something.
var (
	key = []byte("mySuperSecretKeyLol")
)

var (
	errMissingMetadata = status.Errorf(codes.InvalidArgument, "missing metadata")
	errInvalidToken    = status.Errorf(codes.Unauthenticated, "invalid token")
)

// UserCtxKey is type for key to access user value from context. It's currently
// exported for possible outside use.
// NOTE: Must be own type and cannot be type alias because linter warnings with
// context!
type UserCtxKey string

const userKey = "UserKey"

type customClaims struct {
	Username string `json:"un"`
	Label    string `json:"label,omitempty"`
	*jwt.StandardClaims
}

func SetJWTSecret(jwtSecret string) {
	// todo: remove from the log after jwt secret flag is used commonly
	glog.V(3).Infoln("===== USING given JWT secret ====")
	key = []byte(jwtSecret)
}

// User is a helper function to get user from the current ctx as a string.
func User(ctx context.Context) string {
	return ctx.Value(UserCtxKey(userKey)).(string)
}

func NewContextWithUser(ctx context.Context, user string) context.Context {
	return context.WithValue(ctx, UserCtxKey(userKey), user)
}

// BuildJWT builds a signed JWT token from user string. User string can be user
// ID, or DID, or something similar. This function is called to generate a token
// for client. The token is checked with the check function.
func BuildJWT(user string) string {
	return BuildJWTWithLabel(user, "")
}

func BuildJWTWithLabel(user, label string) string {
	const timeValid = 72 * time.Hour

	return BuildJWTWithTime(user, label, timeValid)
}

func BuildJWTWithTime(user, label string, timeValid time.Duration) string {
	claims := &customClaims{
		Username: user,
		Label:    label,
		StandardClaims: &jwt.StandardClaims{
			ExpiresAt: time.Now().Add(timeValid).Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString(key)
	if err != nil {
		return ""
	}
	return ss
}

// check checks the JWT token and writes the Username to Context.
func check(ctx context.Context, ts string) (context.Context, bool) {
	token, err := jwt.ParseWithClaims(ts, &customClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return key, nil
		})

	if err != nil {
		glog.Error(err)
		return ctx, false
	}
	if claims, ok := token.Claims.(*customClaims); ok && token.Valid {
		ctx = context.WithValue(ctx, UserCtxKey(userKey), claims.Username)
	} else {
		glog.Error("no claims in token")
		return ctx, false
	}

	return ctx, token.Valid
}

func validateAndUser(ts string) string {
	token, err := jwt.ParseWithClaims(ts, &customClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return key, nil
		})

	if err != nil {
		glog.Error(err)
		return ""
	}
	if claims, ok := token.Claims.(*customClaims); ok && token.Valid {
		return claims.Username
	}
	glog.Error("no claims in token")
	return ""
}

// valid validates the authorization and returns new ctx with it.
func valid(ctx context.Context, authorization []string) (context.Context, bool) {
	if len(authorization) < 1 {
		glog.Error("no authorization meta data")
		return ctx, false
	}
	prefix := "Bearer "
	token := strings.TrimPrefix(authorization[0], prefix)
	glog.V(13).Infoln("token:", token)
	for _, a := range authorization {
		if strings.HasPrefix(a, prefix) {
			token := strings.TrimPrefix(a, prefix)
			glog.V(10).Infoln("token:", token)
			if validateAndUser(token) != "" {
				return check(ctx, token)
			}
		}
	}
	return ctx, false
}

// ParseValidate validates the JWT token and that cheks it contains the correct
// the Username.
func ParseValidate(user, tokenStr string) bool {
	return user == validateAndUser(tokenStr)
}

// IsValidUser loops thru the authorization and checks that correct user is in
// the Bearer JWT token.
func IsValidUser(user string, authorization []string) bool {
	if len(authorization) < 1 {
		glog.Error("no authorization data")
		return false
	}
	prefix := "Bearer "
	for _, a := range authorization {
		if strings.HasPrefix(a, prefix) {
			token := strings.TrimPrefix(a, prefix)
			glog.V(10).Infoln("token:", token)
			if ParseValidate(user, token) {
				return true
			}
		}
	}
	return false
}

// IsTimeLeft calculates if time is left in token with given duration.
func IsTimeLeft(tokenStr string, delta time.Duration) bool {
	token, err := jwt.ParseWithClaims(tokenStr, &customClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return key, nil
		})

	if err != nil {
		glog.Error(err)
		return false
	}

	if claims, ok := token.Claims.(*customClaims); ok && token.Valid {
		stamp := time.Now().Add(delta).Unix()
		return claims.StandardClaims.VerifyExpiresAt(stamp, false)
	}
	glog.Error("no claims in token")
	return false

}

func TimeLeft(tokenStr string) (t time.Duration, err error) {
	defer err2.Handle(&err)

	p := new(jwt.Parser)
	token, _ := try.To2(p.ParseUnverified(tokenStr, &customClaims{}))

	if claims, ok := token.Claims.(*customClaims); ok {
		stamp := time.Now()
		expTime := time.Unix(claims.ExpiresAt, 0)
		return expTime.Sub(stamp), nil
	}
	glog.Error("no claims in token")
	return 0, errors.New("no claims in token")

}

// EnsureValidToken ensures a valid token exists within a request's metadata. If
// the token is missing or invalid, the interceptor blocks execution of the
// handler and returns an error. Otherwise, the interceptor invokes the unary
// handler.
func EnsureValidToken(ctx context.Context, req interface{},
	info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {

	newCtx, err := CheckTokenValidity(ctx)
	if err != nil {
		glog.Error("NO authorization")
		return nil, err
	}
	// Continue execution of handler after ensuring a valid token.
	return handler(newCtx, req)
}

// EnsureValidTokenStream ensures a valid token exists within a request's metadata. If
// the token is missing or invalid, the interceptor blocks execution of the
// handler and returns an error. Otherwise, the interceptor invokes the unary
// handler.
func EnsureValidTokenStream(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	ctx := ss.Context()
	_, err := CheckTokenValidity(ctx)
	if err != nil {
		return err
	}
	// Continue execution of handler after ensuring a valid token.
	return handler(srv, ss)
}

// CheckTokenValidity check if context includes valid JWT and if so, wraps a new
// one with valid user ID.
func CheckTokenValidity(ctx context.Context) (context.Context, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errMissingMetadata
	}

	newCtx, isValid := valid(ctx, md["authorization"])
	if !isValid {
		glog.Error("NO authorization in the token")
		return newCtx, errInvalidToken
	}
	return newCtx, nil
}

// OauthToken returns our JWT token as an oauth because it helps wrapping it to gRPC
// credentials.
func OauthToken(jwt string) *oauth2.Token {
	return &oauth2.Token{
		AccessToken: jwt,
	}
}
