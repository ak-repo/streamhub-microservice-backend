package postgres

import (
	"context"
	"fmt"

	"github.com/ak-repo/stream-hub/internal/payment_service/domain"
	"github.com/ak-repo/stream-hub/internal/payment_service/port"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PgxRepository struct {
	pool *pgxpool.Pool
}

func NewPgxRepository(pool *pgxpool.Pool) port.Repository {
	return &PgxRepository{pool: pool}
}

func (r *PgxRepository) RecordPaymentHistory(
	ctx context.Context,
	history *domain.PaymentHistory,
) error {

	query := `
		INSERT INTO payment_history (
			id,
			channel_id,
			purchaser_user_id,
			razorpay_order_id,
			razorpay_payment_id,
			amount_paid_cents,
			storage_added_bytes,
			created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7,$8)
	`

	_, err := r.pool.Exec(ctx, query,
		history.ID,
		history.ChannelID,
		history.PurchaserUserID,
		history.RazorpayOrderID,
		history.RazorpayPaymentID,
		history.AmountPaidCents,
		history.StorageAddedBytes,
		history.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to record payment history: %w", err)
	}

	return nil
}

func (r *PgxRepository) GetHistoryByChannel(
	ctx context.Context,
	channelID uuid.UUID,
) ([]*domain.PaymentHistory, error) {

	query := `
		SELECT
			id,
			channel_id,
			purchaser_user_id,
			razorpay_order_id,
			razorpay_payment_id,
			amount_paid_cents,
			storage_added_bytes,
			created_at
		FROM channel_payment_history
		WHERE channel_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, channelID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to query payment history: %w", err)
	}
	defer rows.Close()

	histories := make([]*domain.PaymentHistory, 0)

	for rows.Next() {
		var h domain.PaymentHistory

		if err := rows.Scan(
			&h.ID,
			&h.ChannelID,
			&h.PurchaserUserID,
			&h.RazorpayOrderID,
			&h.RazorpayPaymentID,
			&h.AmountPaidCents,
			&h.StorageAddedBytes,
			&h.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan payment history: %w", err)
		}

		histories = append(histories, &h)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return histories, nil
}
