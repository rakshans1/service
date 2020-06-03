package handlers

import (
	"context"
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/rakshans1/service/internal/platform/auth"
	"github.com/rakshans1/service/internal/platform/web"
	"github.com/rakshans1/service/internal/user"
	"go.opentelemetry.io/otel/api/global"
)

// Users holds handlers for dealing with user.
type Users struct {
	db            *sqlx.DB
	authenticator *auth.Authenticator
}

// Token generates an authentication token for a user. The client must include
// an email and password for the request using HTTP Basic Auth. The user will
// be identified by email and authenticated by their password.
func (u *Users) Token(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := global.Tracer("service").Start(ctx, "handlers.users.token")
	defer span.End()

	v, ok := ctx.Value(web.KeyValues).(*web.Values)
	if !ok {
		return web.NewShutdownError("web value missing from context")
	}

	email, pass, ok := r.BasicAuth()
	if !ok {
		err := errors.New("must provide email and password in Basic auth")
		return web.NewRequestError(err, http.StatusUnauthorized)
	}

	claims, err := user.Authenticate(ctx, u.db, v.Start, email, pass)
	if err != nil {
		switch err {
		case user.ErrAuthenticationFailure:
			return web.NewRequestError(err, http.StatusUnauthorized)
		default:
			return errors.Wrap(err, "authenticating")
		}
	}

	var tkn struct {
		Token string `json:"token"`
	}
	tkn.Token, err = u.authenticator.GenerateToken(claims)
	if err != nil {
		return errors.Wrap(err, "generating token")
	}

	return web.Respond(ctx, w, tkn, http.StatusOK)
}