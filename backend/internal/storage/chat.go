package storage

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

var (
	ErrConversationNotFound = errors.New("conversation not found")
	ErrMessageNotFound      = errors.New("message not found")
	ErrArtifactNotFound     = errors.New("artifact not found")
)

type Conversation struct {
	ID        string    `json:"id"`
	UserID    int64     `json:"user_id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Message struct {
	ID             string    `json:"id"`
	ConversationID string    `json:"conversation_id"`
	Role           string    `json:"role"`
	Content        string    `json:"content"`
	ArtifactID     *string   `json:"artifact_id,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

type Artifact struct {
	ID             string    `json:"id"`
	ConversationID string    `json:"conversation_id"`
	Type           string    `json:"type"`
	Payload        string    `json:"payload"`
	CreatedAt      time.Time `json:"created_at"`
}

func (db *DB) CreateConversation(ctx context.Context, userID int64, title string) (*Conversation, error) {
	var c Conversation
	err := db.pool.QueryRow(ctx,
		`INSERT INTO conversations (user_id, title) VALUES ($1, $2)
		 RETURNING id, user_id, title, created_at, updated_at`,
		userID, title,
	).Scan(&c.ID, &c.UserID, &c.Title, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (db *DB) ListConversations(ctx context.Context, userID int64) ([]Conversation, error) {
	rows, err := db.pool.Query(ctx,
		`SELECT id, user_id, title, created_at, updated_at
		 FROM conversations WHERE user_id = $1
		 ORDER BY updated_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Conversation
	for rows.Next() {
		var c Conversation
		if err := rows.Scan(&c.ID, &c.UserID, &c.Title, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		result = append(result, c)
	}
	return result, rows.Err()
}

func (db *DB) GetConversation(ctx context.Context, id string) (*Conversation, error) {
	var c Conversation
	err := db.pool.QueryRow(ctx,
		`SELECT id, user_id, title, created_at, updated_at
		 FROM conversations WHERE id = $1`,
		id,
	).Scan(&c.ID, &c.UserID, &c.Title, &c.CreatedAt, &c.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrConversationNotFound
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (db *DB) DeleteConversation(ctx context.Context, id string, userID int64) error {
	tag, err := db.pool.Exec(ctx,
		`DELETE FROM conversations WHERE id = $1 AND user_id = $2`,
		id, userID,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrConversationNotFound
	}
	return nil
}

func (db *DB) UpdateConversationTitle(ctx context.Context, id string, title string) error {
	tag, err := db.pool.Exec(ctx,
		`UPDATE conversations SET title = $1, updated_at = $2 WHERE id = $3`,
		title, time.Now(), id,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrConversationNotFound
	}
	return nil
}

func (db *DB) CreateMessage(ctx context.Context, conversationID, role, content string, artifactID *string) (*Message, error) {
	var m Message
	err := db.pool.QueryRow(ctx,
		`INSERT INTO messages (conversation_id, role, content, artifact_id)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, conversation_id, role, content, artifact_id, created_at`,
		conversationID, role, content, artifactID,
	).Scan(&m.ID, &m.ConversationID, &m.Role, &m.Content, &m.ArtifactID, &m.CreatedAt)
	if err != nil {
		return nil, err
	}

	db.pool.Exec(ctx,
		`UPDATE conversations SET updated_at = $1 WHERE id = $2`,
		time.Now(), conversationID,
	)

	return &m, nil
}

func (db *DB) ListMessages(ctx context.Context, conversationID string, limit int, before string) ([]Message, error) {
	if limit <= 0 || limit > 100 {
		limit = 30
	}

	var rows pgx.Rows
	var err error

	if before != "" {
		rows, err = db.pool.Query(ctx,
			`SELECT id, conversation_id, role, content, artifact_id, created_at
			 FROM messages
			 WHERE conversation_id = $1 AND created_at < (
			   SELECT created_at FROM messages WHERE id = $2
			 )
			 ORDER BY created_at DESC
			 LIMIT $3`,
			conversationID, before, limit,
		)
	} else {
		rows, err = db.pool.Query(ctx,
			`SELECT id, conversation_id, role, content, artifact_id, created_at
			 FROM messages
			 WHERE conversation_id = $1
			 ORDER BY created_at DESC
			 LIMIT $2`,
			conversationID, limit,
		)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]Message, 0, limit)
	for rows.Next() {
		var m Message
		if err := rows.Scan(&m.ID, &m.ConversationID, &m.Role, &m.Content, &m.ArtifactID, &m.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, m)
	}
	return result, rows.Err()
}

func (db *DB) CreateArtifact(ctx context.Context, conversationID, artifactType, payload string) (*Artifact, error) {
	var a Artifact
	err := db.pool.QueryRow(ctx,
		`INSERT INTO artifacts (conversation_id, type, payload) VALUES ($1, $2, $3::jsonb)
		 RETURNING id, conversation_id, type, payload::text, created_at`,
		conversationID, artifactType, payload,
	).Scan(&a.ID, &a.ConversationID, &a.Type, &a.Payload, &a.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (db *DB) GetArtifact(ctx context.Context, id string) (*Artifact, error) {
	var a Artifact
	err := db.pool.QueryRow(ctx,
		`SELECT id, conversation_id, type, payload::text, created_at
		 FROM artifacts WHERE id = $1`,
		id,
	).Scan(&a.ID, &a.ConversationID, &a.Type, &a.Payload, &a.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrArtifactNotFound
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (db *DB) UpdateConversationTimestamp(ctx context.Context, id string) error {
	_, err := db.pool.Exec(ctx,
		`UPDATE conversations SET updated_at = $1 WHERE id = $2`,
		time.Now(), id,
	)
	return err
}
