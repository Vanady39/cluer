package repositories

import (
	"errors"

	"github.com/Vanady39/cluer/internal/domains"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// PostgreSQL error codes we translate rather than pass through.
const (
	pgUniqueViolation     = "23505"
	pgRestrictViolation   = "23001" // raised by the immutability triggers
	pgForeignKeyViolation = "23503"
)

// wrap turns a driver error into something the HTTP layer can answer with.
//
// The immutability triggers raise 23001, and the partial unique indexes raise
// 23505 when two publishes race. Both are meaningful outcomes, not internal
// failures, and returning a raw driver message for either would tell the user
// nothing and leak the schema at the same time.
func wrap(err error, sentinel error, ref string, op domains.RepoOperation) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return domains.NotFound(sentinel, ref)
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case pgRestrictViolation:
			return &domains.LogicError{
				Err:   domains.ErrVersionImmutable,
				Stage: "immutability guard",
				Code:  409,
			}
		case pgUniqueViolation:
			return &domains.RepositoryError{
				Err: err, EntityRef: ref, Operation: op, Reason: domains.AlreadyExists,
			}
		case pgForeignKeyViolation:
			return &domains.RepositoryError{
				Err: err, EntityRef: ref, Operation: op, Reason: domains.InvalidReference,
			}
		}
	}

	reason := domains.QueryError
	if op != domains.Retrieve {
		reason = domains.ExecError
	}
	return &domains.RepositoryError{Err: err, EntityRef: ref, Operation: op, Reason: reason}
}
