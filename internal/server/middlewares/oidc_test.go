package middlewares

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Vanady39/cluer/internal/auth"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

const (
	testIssuer   = "https://auth.example.com/application/o/cluer/"
	testClientID = "cluer-admin"
	testSubject  = "ak-subject-1"
	adminGroup   = "cluer-admins"
)

// echoKeySet stands in for the IdP's published keys. It accepts every
// signature and returns the payload exactly as it appears in the token, which
// is what go-oidc compares against the payload it parsed itself.
//
// Signature verification is go-oidc's job and is covered by its own tests;
// what needs covering here is the middleware wrapped around it. Faking the key
// set keeps these tests off the network and out of RSA key generation.
type echoKeySet struct {
	err error
}

func (k echoKeySet) VerifySignature(_ context.Context, jwt string) ([]byte, error) {
	if k.err != nil {
		return nil, k.err
	}

	parts := strings.Split(jwt, ".")
	if len(parts) != 3 {
		return nil, errors.New("malformed jwt")
	}

	return base64.RawURLEncoding.DecodeString(parts[1])
}

func testVerifier(keySet oidc.KeySet) *oidc.IDTokenVerifier {
	return oidc.NewVerifier(testIssuer, keySet, &oidc.Config{ClientID: testClientID})
}

// token assembles a JWT-shaped string around the given claims. The signature
// segment is inert but has to be decodable base64url, because a token that
// fails to parse would exercise a different branch than the one under test.
func token(t *testing.T, claims map[string]any) string {
	t.Helper()

	payload, err := json.Marshal(claims)
	require.NoError(t, err)

	return strings.Join([]string{
		base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"RS256"}`)),
		base64.RawURLEncoding.EncodeToString(payload),
		base64.RawURLEncoding.EncodeToString([]byte("signature")),
	}, ".")
}

func validClaims() map[string]any {
	return map[string]any{
		"iss":            testIssuer,
		"aud":            testClientID,
		"sub":            testSubject,
		"exp":            time.Now().Add(time.Hour).Unix(),
		"iat":            time.Now().Unix(),
		"email":          "admin@example.com",
		"email_verified": true,
		"name":           "Admin User",
		"groups":         []string{adminGroup},
	}
}

// newTestEngine mounts the middleware behind the real ErrorHandler, because
// the middleware itself only records an error and aborts — the status code and
// body are produced downstream. Asserting on codes without ErrorHandler in the
// chain would assert on nothing.
func newTestEngine(handlers ...gin.HandlerFunc) *gin.Engine {
	logger := zerolog.Nop()

	engine := gin.New()
	engine.Use(ErrorHandler(&logger))
	engine.Use(handlers...)

	// Reads the claims back out of the context, so a 200 also proves they
	// survived the gob round trip rather than merely that nothing rejected the
	// request.
	engine.GET("/probe", func(c *gin.Context) {
		claims, err := auth.FromGin(c)
		if err != nil {
			c.String(http.StatusOK, "no-claims")
			return
		}
		c.String(http.StatusOK, claims.Subject)
	})

	return engine
}

func do(engine *gin.Engine, header string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/probe", nil)
	if header != "" {
		req.Header.Set("Authorization", header)
	}

	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	return rec
}

func TestOIDCAuthRejects(t *testing.T) {
	expired := validClaims()
	expired["exp"] = time.Now().Add(-time.Minute).Unix()

	foreignIssuer := validClaims()
	foreignIssuer["iss"] = "https://auth.attacker.example/application/o/cluer/"

	foreignAudience := validClaims()
	foreignAudience["aud"] = "some-other-client"

	tests := []struct {
		name   string
		header func(t *testing.T) string
		keySet oidc.KeySet
	}{
		{
			name:   "no header at all",
			header: func(*testing.T) string { return "" },
		},
		{
			name:   "wrong scheme",
			header: func(t *testing.T) string { return "Token " + token(t, validClaims()) },
		},
		{
			name:   "scheme without a token",
			header: func(*testing.T) string { return "Bearer " },
		},
		{
			name:   "not a jwt",
			header: func(*testing.T) string { return "Bearer not-a-token" },
		},
		{
			name:   "expired",
			header: func(t *testing.T) string { return "Bearer " + token(t, expired) },
		},
		{
			name:   "issued by someone else",
			header: func(t *testing.T) string { return "Bearer " + token(t, foreignIssuer) },
		},
		{
			name:   "addressed to another audience",
			header: func(t *testing.T) string { return "Bearer " + token(t, foreignAudience) },
		},
		{
			name:   "signature does not check out",
			header: func(t *testing.T) string { return "Bearer " + token(t, validClaims()) },
			keySet: echoKeySet{err: errors.New("signature mismatch")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keySet := tt.keySet
			if keySet == nil {
				keySet = echoKeySet{}
			}

			engine := newTestEngine(OIDCAuth(testVerifier(keySet)))
			rec := do(engine, tt.header(t))

			require.Equal(t, http.StatusUnauthorized, rec.Code)
		})
	}
}

func TestOIDCAuthAcceptsAndExposesClaims(t *testing.T) {
	engine := newTestEngine(OIDCAuth(testVerifier(echoKeySet{})))

	rec := do(engine, "Bearer "+token(t, validClaims()))

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, testSubject, rec.Body.String())
}

// RFC 6750 makes the scheme case-insensitive; a client sending "bearer" is
// within spec and must not be turned away.
func TestOIDCAuthAcceptsLowercaseScheme(t *testing.T) {
	engine := newTestEngine(OIDCAuth(testVerifier(echoKeySet{})))

	rec := do(engine, "bearer "+token(t, validClaims()))

	require.Equal(t, http.StatusOK, rec.Code)
}

func TestRequireGroups(t *testing.T) {
	memberOfSeveral := validClaims()
	memberOfSeveral["groups"] = []string{"authentik Admins", adminGroup}

	noGroups := validClaims()
	delete(noGroups, "groups")

	outsider := validClaims()
	outsider["groups"] = []string{"some-other-team"}

	// Group names are compared exactly: a case mismatch is a configuration
	// error in the IdP, and silently accepting it here would make the config
	// mean something different from what it says.
	wrongCase := validClaims()
	wrongCase["groups"] = []string{strings.ToUpper(adminGroup)}

	tests := []struct {
		name   string
		claims map[string]any
		want   int
	}{
		{name: "carries the admin group", claims: validClaims(), want: http.StatusOK},
		{name: "carries it among others", claims: memberOfSeveral, want: http.StatusOK},
		{name: "carries a different group", claims: outsider, want: http.StatusForbidden},
		{name: "carries no groups", claims: noGroups, want: http.StatusForbidden},
		{name: "differs in case", claims: wrongCase, want: http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := newTestEngine(
				OIDCAuth(testVerifier(echoKeySet{})),
				RequireGroups([]string{adminGroup}),
			)

			rec := do(engine, "Bearer "+token(t, tt.claims))

			require.Equal(t, tt.want, rec.Code)
		})
	}
}

// Reaching RequireGroups without OIDCAuth in front of it means the route was
// wired up wrong. That is ours to fix, not the caller's to work around, so it
// answers 500 rather than 401 or 403.
func TestRequireGroupsWithoutAuthentication(t *testing.T) {
	engine := newTestEngine(RequireGroups([]string{adminGroup}))

	rec := do(engine, "Bearer "+token(t, validClaims()))

	require.Equal(t, http.StatusInternalServerError, rec.Code)
}
