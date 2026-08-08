package repositories

import (
	"errors"

	"github.com/Vanady39/cluer/internal/domains"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	pgUniqueViolation     = "23505"
	pgRestrictViolation   = "23001"
	pgForeignKeyViolation = "23503"
)

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
