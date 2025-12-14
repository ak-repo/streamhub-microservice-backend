package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/ak-repo/stream-hub/internal/payment_service/domain"
	"github.com/ak-repo/stream-hub/internal/payment_service/port"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type paymentRepo struct {
	pool *pgxpool.Pool
}

func NewPaymentRepo(pool *pgxpool.Pool) port.Repository {
	return &paymentRepo{pool: pool}
}

func (r *paymentRepo) RecordPaymentHistory(
	ctx context.Context,
	history *domain.PaymentHistory,
) error {

	query := `
		INSERT INTO channel_subscriptions (
			id,
			channel_id,
			plan_id,
			razorpay_order_id,
			razorpay_payment_id,
			created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.pool.Exec(ctx, query,
		history.ID,
		history.ChannelID,
		history.PlanID,
		history.RazorpayOrderID,
		history.RazorpayPaymentID,
		history.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to record payment history: %w", err)
	}

	return nil
}

func (r *paymentRepo) GetHistoryByChannel(
	ctx context.Context,
	channelID uuid.UUID,
) ([]*domain.PaymentHistory, error) {

	// TODO errors

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

func (r *paymentRepo) ListSubscriptionPlans(ctx context.Context) ([]*domain.SubscriptionPlan, error) {

	raws, err := r.pool.Query(ctx, `SELECT id, name, storage_limit_mb, price_inr,  duration_days FROM  channel_storage_plans`)
	if err != nil {
		return nil, err
	}
	defer raws.Close()

	var plans []*domain.SubscriptionPlan

	for raws.Next() {
		p := &domain.SubscriptionPlan{}
		if err := raws.Scan(&p.ID, &p.Name, &p.StorageLimitMB, &p.PriceINR, &p.DurationDays); err != nil {
			return nil, err
		}

		plans = append(plans, p)

	}
	return plans, raws.Err()
}

func (r *paymentRepo) PlanByID(ctx context.Context, planID string) (*domain.SubscriptionPlan, error) {
	raw := r.pool.QueryRow(ctx, `SELECT id, name, storage_limit_mb, price_inr,  duration_days FROM  channel_storage_plans WHERE id = $1`, planID)

	plan := &domain.SubscriptionPlan{}
	if err := raw.Scan(&plan.ID, &plan.Name, &plan.StorageLimitMB, &plan.PriceINR, &plan.DurationDays); err != nil {
		return nil, err
	}

	return plan, nil

}

func (r *paymentRepo) ChannelPlanID(ctx context.Context, channelID string) (string, error) {
	var planID sql.NullString

	row := r.pool.QueryRow(
		ctx,
		`SELECT active_plan_id FROM channels WHERE id = $1`,
		channelID,
	)

	if err := row.Scan(&planID); err != nil {
		// if errors.Is(err, pgx.ErrNoRows) {
		// 	return "", ErrChannelNotFound
		// }
		return "", err
	}

	// if !planID.Valid {
	// 	return "",fmt.Errorf("no active plan")
	// }

	log.Println("plan id: ", planID.String)
	return planID.String, nil
}
