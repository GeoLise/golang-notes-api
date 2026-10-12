-- +goose Up
CREATE TABLE notes (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    name VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_notes_user
        FOREIGN KEY (user_id)
        REFERENCES users (id)
        ON DELETE RESTRICT ON UPDATE CASCADE
);

-- +goose Down
DROP TABLE notes;
