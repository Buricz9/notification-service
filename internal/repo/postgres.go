// internal/repo/postgres.go
package repo

import (
	"context"
	"time"

	"github.com/Buricz9/notification-service/internal/model"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/streadway/amqp"
)

type PostgresRepo struct {
	db *sqlx.DB
	mq *amqp.Channel
}

func NewPostgresRepo(db *sqlx.DB, mq *amqp.Channel) NotificationRepository {
	return &PostgresRepo{db: db, mq: mq}
}

func (r *PostgresRepo) Create(ctx context.Context, n *model.Notification) (uuid.UUID, error) {
	// Wygeneruj ID, jeśli nie ma
	if n.ID == uuid.Nil {
		n.ID = uuid.New()
	}
	// Dodajemy send_at, timezone i priority do INSERT-a
	query := `
        INSERT INTO notifications 
            (id, user_id, channel, payload, status, send_at, timezone, priority)
        VALUES 
            (:id, :user_id, :channel, :payload, :status, :send_at, :timezone, :priority)
    `
	if _, err := r.db.NamedExecContext(ctx, query, n); err != nil {
		return uuid.Nil, err
	}

	// Wypchnij zdarzenie do MQ
	body := n.ID.String()
	if err := r.mq.Publish(
		"",              // exchange
		"notifications", // routing key
		false, false,
		amqp.Publishing{ContentType: "text/plain", Body: []byte(body)},
	); err != nil {
		return n.ID, err
	}

	return n.ID, nil
}

func (r *PostgresRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Notification, error) {
	var n model.Notification
	err := r.db.GetContext(ctx, &n, "SELECT * FROM notifications WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &n, nil
}

func (r *PostgresRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status string, retry int, errMsg *string) error {
	if errMsg == nil {
		// usuń stare error
		_, err := r.db.ExecContext(ctx,
			`UPDATE notifications SET status=$1, retry_count=$2, error=NULL, updated_at=NOW() WHERE id=$3`,
			status, retry, id)
		return err
	}
	_, err := r.db.ExecContext(ctx,
		`UPDATE notifications SET status=$1, retry_count=$2, error=$3, updated_at=NOW() WHERE id=$4`,
		status, retry, *errMsg, id)
	return err
}

// PostgresRepo implementuje:
func (r *PostgresRepo) ListPendingBefore(ctx context.Context, before time.Time) ([]model.Notification, error) {
	var out []model.Notification
	err := r.db.SelectContext(ctx, &out, `
        SELECT * FROM notifications
         WHERE status = 'pending' AND send_at <= $1
         ORDER BY priority DESC, send_at ASC
    `, before)
	return out, err
}

func (r *PostgresRepo) MarkScheduled(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `
        UPDATE notifications
           SET status = 'scheduled', updated_at = NOW()
         WHERE id = $1
    `, id)
	return err
}

func (r *PostgresRepo) ListByStatus(ctx context.Context, status string) ([]model.Notification, error) {
	var out []model.Notification
	err := r.db.SelectContext(ctx, &out,
		`SELECT * FROM notifications WHERE status = $1 ORDER BY send_at ASC`, status)
	return out, err
}

func (r *PostgresRepo) Stats(ctx context.Context, from, to time.Time) (map[string]int, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT status, COUNT(*) 
           FROM notifications 
          WHERE created_at BETWEEN $1 AND $2
          GROUP BY status`, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := make(map[string]int)
	for rows.Next() {
		var status string
		var cnt int
		if err := rows.Scan(&status, &cnt); err != nil {
			return nil, err
		}
		res[status] = cnt
	}
	return res, nil
}

func (r *PostgresRepo) ForceSend(ctx context.Context, id uuid.UUID) (bool, error) {
	result, err := r.db.ExecContext(ctx,
		`UPDATE notifications 
            SET send_at = NOW(), status = 'pending', updated_at = NOW()
          WHERE id = $1 AND status = 'pending'`, id)
	if err != nil {
		return false, err
	}
	n, _ := result.RowsAffected()
	return n > 0, nil
}

func (r *PostgresRepo) Cancel(ctx context.Context, id uuid.UUID) (bool, error) {
	result, err := r.db.ExecContext(ctx,
		`UPDATE notifications 
            SET status = 'failed', updated_at = NOW()
          WHERE id = $1 AND status = 'pending'`, id)
	if err != nil {
		return false, err
	}
	n, _ := result.RowsAffected()
	return n > 0, nil
}
