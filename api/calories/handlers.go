package calories

import (
	"context"
	"net/http"
	"slices"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/scottapow/scottapow/api/auth"
	db "github.com/scottapow/scottapow/data"
)

type CaloriesService struct {
	Auth *auth.AuthProvider
}

func NewCaloriesService(auth *auth.AuthProvider) *CaloriesService {
	return &CaloriesService{
		Auth: auth,
	}
}

func (s *CaloriesService) GetAllCaloriesData(ctx context.Context, userId string) ([]db.CalorieEntriesModal, error) {
	tx, err := s.Auth.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}

	var id pgtype.UUID
	err = tx.QueryRow(ctx, `
		SELECT id FROM calories WHERE user_id = $1;
	`, userId).Scan(&id)

	if err != nil || id.Valid == false {
		tx.Rollback(ctx)
		return nil, err
	}

	rows, err := tx.Query(ctx, `
		SELECT id, calories_id, amount, consumed_at FROM calorie_entries WHERE calories_id = $1;
	`, id.String())
	defer rows.Close()

	if err != nil {
		tx.Rollback(ctx)
		return nil, err
	}

	var entries []db.CalorieEntriesModal
	for rows.Next() {
		var entry db.CalorieEntriesModal
		rows.Scan(&entry.Id, &entry.Calories_id, &entry.Amount, &entry.Consumed_at)
		entries = append(entries, entry)
	}

	tx.Commit(ctx)
	return entries, nil
}

func (s *CaloriesService) AddCaloriesEntry(ctx context.Context, caloriesId string, amount int16, timestamp string) error {
	_, err := s.Auth.DB.Query(ctx, `
		INSERT INTO calorie_entries (calories_id, amount, consumed_at) 
		VALUES ($1, $2, $3);
	`, caloriesId, amount, timestamp)

	return err
}

func (s *CaloriesService) WithPermission(permission string, fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, err := s.Auth.GetUserClaims(r)
		if err != nil || !slices.Contains(claims.Permissions, permission) {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}

		fn(w, r)
	}
}
