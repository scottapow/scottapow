package dumps

import (
	"context"
	"net/http"
	"slices"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
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

func (s *DumpsService) GetAllDumpsData(ctx context.Context, userId string) ([]db.DumpEntriesModal, error) {
	tx, err := s.Auth.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}

	var id pgtype.UUID
	err = tx.QueryRow(ctx, `
		SELECT id FROM dumps WHERE user_id = $1;
	`, userId).Scan(&id)

	if err != nil || id.Valid == false {
		tx.Rollback(ctx)
		return nil, err
	}

	rows, err := tx.Query(ctx, `
		SELECT id, dumps_id, amount, occurred_at FROM dump_entries WHERE dumps_id = $1;
	`, id.String())

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
