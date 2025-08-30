package dumps

import (
	"context"
	"errors"
	"net/http"
	"slices"

	"github.com/jackc/pgx/v5"
	"github.com/scottapow/scottapow/api/auth"
	db "github.com/scottapow/scottapow/data"
)

type DumpsService struct {
	Auth *auth.AuthProvider
}

func NewDumpsService(auth *auth.AuthProvider) *DumpsService {
	return &DumpsService{
		Auth: auth,
	}
}

func (s *DumpsService) GetAllDumps(ctx context.Context, userId string) ([]db.DumpsModel, error) {
	tx, err := s.Auth.DB.Begin(ctx)
	if err != nil {
		return nil, err
	}

	var dumps []db.DumpsModel
	rows, err := tx.Query(ctx, `
		SELECT * FROM dumps WHERE user_id = $1;
	`, userId)

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		d := db.DumpsModel{}
		err := rows.Scan(&d.Id, &d.User_id, &d.Created_at, &d.Updated_at, &d.Description)
		if err != nil {
			return nil, err
		}
		dumps = append(dumps, d)
	}

	return dumps, nil
}

func (s *DumpsService) GetDump(ctx context.Context, userId string, dumpId string) (db.DumpsModel, error) {
	var dump = db.DumpsModel{}

	if dumpId == "" {
		return dump, errors.New("No dumpId provided")
	}

	tx, err := s.Auth.DB.Begin(ctx)
	if err != nil {
		return dump, err
	}

	err = tx.QueryRow(ctx, `
		SELECT id, user_id, created_at, updated_at, description FROM dumps WHERE user_id = $1;
	`, userId).Scan(&dump.Id, &dump.User_id, &dump.Created_at, &dump.Updated_at, &dump.Description)

	if err != nil {
		return dump, err
	}

	return dump, nil
}

func (s *DumpsService) GetDumpsData(ctx context.Context, dumpId string, userId string) ([]db.DumpEntriesModal, error) {
	if dumpId == "" {
		return nil, errors.New("No dumpId provided")
	}

	tx, err := s.Auth.DB.BeginTx(ctx, pgx.TxOptions{})

	if err != nil {
		tx.Rollback(ctx)
		return nil, err
	}

	// check if the current user is the creator of the dump
	// only creators are currently allowed access
	tag, err := tx.Exec(ctx, `SELECT EXISTS(SELECT 1 from dumps WHERE user_id = $1 AND id = $2)`, userId, dumpId)

	if err != nil {
		tx.Rollback(ctx)
		return nil, err
	}

	if tag.RowsAffected() == 0 {
		tx.Rollback(ctx)
		return nil, errors.New("This dump cannot be accessed by this user")
	}

	rows, err := tx.Query(ctx, `
		SELECT id, dumps_id, amount, occurred_at FROM dump_entries WHERE dumps_id = $1;
	`, dumpId)

	if err != nil {
		tx.Rollback(ctx)
		return nil, err
	}

	defer rows.Close()

	var entries []db.DumpEntriesModal
	for rows.Next() {
		var entry db.DumpEntriesModal
		rows.Scan(&entry.Id, &entry.Dumps_id, &entry.Amount, &entry.Occurred_at)
		entries = append(entries, entry)
	}

	tx.Commit(ctx)
	return entries, nil
}

func (s *DumpsService) AddDump(ctx context.Context, userId string, desc string) (string, error) {
	var id string
	err := s.Auth.DB.QueryRow(ctx, `
		INSERT INTO dumps (id, user_id, description) 
		VALUES (gen_random_uuid(), $1, $2)
		RETURNING id;
	`, userId, desc,
	).Scan(&id)

	return id, err
}

func (s *DumpsService) AddDumpsEntry(ctx context.Context, dumpsId string, amount int16, timestamp string) error {
	_, err := s.Auth.DB.Query(ctx, `
		INSERT INTO dump_entries (dumps_id, amount, occurred_at) 
		VALUES ($1, $2, $3);
	`, dumpsId, amount, timestamp)

	return err
}

func (s *DumpsService) WithPermission(permission string, fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.Auth == nil {
			http.Error(w, "authentication provider is not initialized", http.StatusInternalServerError)
			return
		}

		claims, err := s.Auth.GetUserClaims(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusFailedDependency)
			return
		}

		if !slices.Contains(claims.Permissions, permission) {
			http.Error(w, "Not Allowed", http.StatusForbidden)
			return
		}

		fn(w, r)
	}
}
