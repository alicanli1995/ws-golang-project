package dbrepo

import (
	"context"
	"golang-vigilate-project/internal/models"
	"time"
)

func (m *postgresDBRepo) CreateSession(params models.CreateSessionsParams) (models.Session, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `INSERT INTO sessions (id,email,refresh_token,user_agent,client_ip,is_blocked,expires_at)
    	VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id,email,refresh_token,user_agent,client_ip,is_blocked,expires_at,created_at`

	row := m.DB.QueryRowContext(ctx, stmt, params.ID, params.Email, params.RefreshToken, params.UserAgent, params.ClientIp, params.IsBlocked, params.ExpiresAt)
	var session models.Session

	err := row.Scan(&session.ID, &session.Email, &session.RefreshToken, &session.UserAgent, &session.ClientIp, &session.IsBlocked, &session.ExpiresAt, &session.CreatedAt)

	if err != nil {
		return session, err
	}

	return session, nil

}
