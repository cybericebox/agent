// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0

package postgres

import (
	"net/netip"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type Laboratory struct {
	ID        uuid.UUID          `json:"id"`
	GroupID   uuid.UUID          `json:"group_id"`
	Cidr      netip.Prefix       `json:"cidr"`
	UpdatedAt pgtype.Timestamptz `json:"updated_at"`
	CreatedAt time.Time          `json:"created_at"`
}
