package groups

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"

	"github.com/Fista6k/disClone/internal"
	"github.com/Fista6k/disClone/internal/domain"
)

type IGroupRepository interface {
	CreateGroup(ctx context.Context, group *Group) error
	GetMyGroups(ctx context.Context, userID int64) ([]Group, error)
	AddNewMembers(ctx context.Context, group_id int64, members_ids []int64) error
	GetGroupByID(ctx context.Context, groupID int64) (*Group, error)
	GetMembersByGroup(ctx context.Context, groupID int64) ([]GroupMember, error)
	DeleteMember(ctx context.Context, groupID int64, userID int64) error
	GetMessageHistory(ctx context.Context, groupID int64) ([]GroupMessage, error)
	CreateGroupMessage(ctx context.Context, message *GroupMessage) (*GroupMessage, error)
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

func (r *GroupRepository) GetMyGroups(ctx context.Context, userID int64) ([]Group, error) {
	query := `
		SELECT g.id, g.name, g.owner_id, g.created_at
		FROM groups g
		JOIN group_members gm ON gm.group_id = g.id
		WHERE gm.user_id = $1
		ORDER BY g.created_at DESC;
	`

	rows, err := r.storage.DB.QueryContext(ctx, query, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			slog.Error("failed to close rows", "error", err)
		}
	}()

	groups := []Group{}
	for rows.Next() {
		var group Group

		err = rows.Scan(&group.ID, &group.Name, &group.OwnerID, &group.CreatedAt)
		if err != nil {
			return nil, err
		}

		groups = append(groups, group)
	}

	if err := rows.Err(); err != nil {
		return nil, err
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
					return domain.ErrUserNotFound
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

func (r *GroupRepository) GetGroupByID(ctx context.Context, groupID int64) (*Group, error) {
	query := `
		SELECT id, name, owner_id, created_at
		FROM groups
		WHERE id = $1;
	`

	var group Group

	err := r.storage.DB.QueryRowContext(ctx, query, groupID).Scan(&group.ID, &group.Name, &group.OwnerID, &group.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrGroupNotFound
		}
		return nil, err
	}

	return &group, nil
}

func (r *GroupRepository) GetMembersByGroup(ctx context.Context, groupID int64) ([]GroupMember, error) {
	query := `
		SELECT group_id, user_id, joined_at
		FROM group_members
		WHERE group_id = $1;
	`

	rows, err := r.storage.DB.QueryContext(ctx, query, groupID)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := rows.Close(); err != nil {
			slog.Error("failed to close rows", "error", err)
		}
	}()

	var members []GroupMember

	for rows.Next() {
		var member GroupMember

		err := rows.Scan(&member.groupID, &member.userID, &member.joinedAt)
		if err != nil {
			return nil, err
		}

		members = append(members, member)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return members, nil
}

func (r *GroupRepository) DeleteMember(ctx context.Context, groupID int64, userID int64) error {
	query := `
		DELETE FROM group_members
		WHERE group_id = $1 AND user_id = $2;
	`

	result, err := r.storage.DB.ExecContext(ctx, query, groupID, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return domain.ErrNotGroupMember
	}

	return nil
}

func (r *GroupRepository) GetMessageHistory(ctx context.Context, groupID int64) ([]GroupMessage, error) {
	query := `
		SELECT id, group_id, author_id, content, created_at
		FROM group_messages
		WHERE group_id = $1;
	`

	rows, err := r.storage.DB.QueryContext(ctx, query, groupID)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := rows.Close(); err != nil {
			slog.Error("failed to close rows", "error", err)
		}
	}()

	messages := make([]GroupMessage, 0)
	for rows.Next() {
		var message GroupMessage

		err = rows.Scan(&message.ID, &message.GroupID, &message.AuthorID, &message.Content, &message.CreatedAt)
		if err != nil {
			return nil, err
		}

		messages = append(messages, message)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return messages, nil
}

func (r *GroupRepository) CreateGroupMessage(ctx context.Context, message *GroupMessage) (*GroupMessage, error) {
	query := `
		INSERT INTO group_messages (group_id, author_id, content, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id;
	`

	err := r.storage.DB.QueryRowContext(ctx, query, message.GroupID, message.AuthorID, message.Content, message.CreatedAt).Scan(&message.ID)
	if err != nil {
		return nil, err
	}

	return message, nil
}
