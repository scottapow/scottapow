package db

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type PGUserModel struct {
	Id                  pgtype.UUID
	Email               pgtype.Text
	Password            pgtype.Text
	Created_at          pgtype.Timestamp
	Updated_at          pgtype.Timestamp
	Login_at            pgtype.Timestamp
	Firstname           pgtype.Text
	Surname             pgtype.Text
	AvatarURL           pgtype.Text
	Oauth_provider      pgtype.Text
	Oauth_provider_id   pgtype.Text
	Oauth_access_token  pgtype.Text
	Oauth_refresh_token pgtype.Text
	Oauth_expires_at    pgtype.Timestamp
}
type UserModel struct {
	Id                  string
	Email               string
	Password            string
	Created_at          time.Time
	Updated_at          time.Time
	Login_at            time.Time
	Firstname           string
	Surname             string
	AvatarURL           string
	Oauth_provider      string
	Oauth_provider_id   string
	Oauth_access_token  string
	Oauth_refresh_token string
	Oauth_expires_at    time.Time
}

type PGPermissionModel struct {
	Id         pgtype.UUID
	User_id    pgtype.UUID
	Permission pgtype.Text
	Created_at pgtype.Timestamp
}
type PermissionModel struct {
	Id         string
	User_id    string
	Permission string
	Created_at time.Time
}

type PGDumpsModel struct {
	Id          pgtype.UUID
	User_id     pgtype.UUID
	Created_at  pgtype.Timestamp
	Updated_at  pgtype.Timestamp
	Description pgtype.Text
}
type DumpsModel struct {
	Id          string
	User_id     string
	Created_at  time.Time
	Updated_at  time.Time
	Description string
}

type PGDumpEntriesModal struct {
	Id          pgtype.UUID
	Dumps_id    pgtype.UUID
	Amount      pgtype.Int2
	Occurred_at pgtype.Timestamptz
}
type DumpEntriesModal struct {
	Id          string
	Dumps_id    string
	Amount      int16
	Occurred_at time.Time
}
