package sqlite

import (
	"context"
	"strings"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/encoding/protojson"

	storepb "github.com/pixb/memos-store/proto/gen/store"
	"github.com/pixb/memos-store/store"
)

func (d *DB) CreateActivity(ctx context.Context, create *store.Activity) (*store.Activity, error) {
	payloadString := "{}"
	if create.Payload != nil {
		bytes, err := protojson.Marshal(create.Payload)
		if err != nil {
			return nil, errors.Wrap(err, "failed to marshal activity payload")
		}
		payloadString = string(bytes)
	}

	fields := []string{"`creator_id`", "`type`", "`level`", "`payload`"}
	placeholder := []string{"?", "?", "?", "?"}
	args := []any{create.CreatorID, create.Type.String(), create.Level.String(), payloadString}

	stmt := "INSERT INTO activity (" + strings.Join(fields, ", ") + ") VALUES (" + strings.Join(placeholder, ", ") + ") RETURNING `id`, `created_ts`"
	if err := d.db.QueryRowContext(ctx, stmt, args...).Scan(
		&create.ID,
		&create.CreatedTs,
	); err != nil {
		return nil, err
	}

	return create, nil
}

func (d *DB) ListActivities(ctx context.Context, find *store.FindActivity) ([]*store.Activity, error) {
	where, args := []string{"1 = 1"}, []any{}
	if find.ID != nil {
		where, args = append(where, "`id` = ?"), append(args, *find.ID)
	}
	if find.Type != nil {
		where, args = append(where, "`type` = ?"), append(args, find.Type.String())
	}

	query := "SELECT `id`, `creator_id`, `type`, `level`, `payload`, `created_ts` FROM `activity` WHERE " + strings.Join(where, " AND ") + " ORDER BY `created_ts` DESC"
	rows, err := d.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []*store.Activity{}
	for rows.Next() {
		activity := &store.Activity{}
		var payloadBytes []byte
		if err := rows.Scan(
			&activity.ID,
			&activity.CreatorID,
			&activity.Type,
			&activity.Level,
			&payloadBytes,
			&activity.CreatedTs,
		); err != nil {
			return nil, err
		}

		payload := &storepb.ActivityPayload{}
		if err := protojsonUnmarshaler.Unmarshal(payloadBytes, payload); err != nil {
			return nil, err
		}
		activity.Payload = payload
		list = append(list, activity)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return list, nil
}
