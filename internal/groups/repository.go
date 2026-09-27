package groups

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/Fista6k/disClone/internal"
)

type IGroupRepository interface {
	CreateGroup(ctx context.Context, group *Group) error
	GetGroups(ctx context.Context, owner_id int64) ([]Group, error)
	AddNewMembers(ctx context.Context, group_id int64, members_ids []int64) error
}

type GroupRepository struct {
	storage *internal.Storage
}

func NewGroupRepository(storage *internal.Storage) *GroupRepository {
	return &GroupRepository{
		storage: storage,
	}
}

func (r *GroupRepository) CreateGroup(ctx context.Context, group *Group) error {
	return internal.WithTx(ctx, r.storage.DB, func(tx *sql.Tx) error {
		query := `
			INSERT INTO groups (name, owner_id, created_at)
			VALUES ($1, $2, $3)
			RETURNING id;
		`

		err := tx.QueryRowContext(ctx, query, group.Name, group.OwnerID, group.CreatedAt).Scan(&group.ID)
		if err != nil {
			return err
		}

		query = `
			INSERT INTO group_members (group_id, user_id, joined_at)
			VALUES ($1, $2, $3);
		`

		_, err = tx.ExecContext(ctx, query, group.ID, group.OwnerID, group.CreatedAt)
		return err
	})
}

func (r *GroupRepository) GetGroups(ctx context.Context, owner_id int64) ([]Group, error) {
	query := `
		SELECT id, name, owner_id, created_at
		FROM groups
		WHERE owner_id = $1;
	`

	rows, err := r.storage.DB.QueryContext(ctx, query, owner_id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("Not found messages to this user")
		}

		return nil, err
	}
	defer rows.Close()

	groups := []Group{}
	for rows.Next() {
		var group Group

		err = rows.Scan(&group.ID, &group.Name, &group.OwnerID, &group.CreatedAt)
		if err != nil {
			return nil, err
		}

		groups = append(groups, group)
	}

	return groups, nil
}

func (r *GroupRepository) AddNewMembers(ctx context.Context, group_id int64, members_ids []int64) error {
	return internal.WithTx(ctx, r.storage.DB, func(tx *sql.Tx) error {
		now := time.Now()
		for _, member_id := range members_ids {
			query := `
				SELECT id
				FROM users
				WHERE id = $1;
			`

			var id int64
			err := tx.QueryRowContext(ctx, query, member_id).Scan(&id)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return errors.New("user not found")
				}
				return err
			}

			query = `
				INSERT INTO group_members (group_id, user_id, joined_at)
				VALUES ($1, $2, $3)
			`

			_, err = tx.ExecContext(ctx, query, group_id, member_id, now)
			if err != nil {
				return err
			}
		}

		return nil
	})
}
