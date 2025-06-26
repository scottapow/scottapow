package db

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type UserModel struct {
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

type PermissionModel struct {
	Id         pgtype.UUID
	User_id    pgtype.UUID
	Permission pgtype.Text
	Created_at pgtype.Timestamp
}

type DumpsModel struct {
	Id          pgtype.UUID
	User_id     pgtype.UUID
	Created_at  pgtype.Timestamp
	Updated_at  pgtype.Timestamp
	Description pgtype.Text
}

type DumpEntriesModal struct {
	Id          pgtype.UUID
	Dumps_id    pgtype.UUID
	Amount      pgtype.Int2
	Occurred_at pgtype.Timestamptz
}
