CREATE TABLE IF NOT EXISTS group_members (
    group_id INT NOT NULL REFERENCES groups(id),
    user_id INT NOT NULL REFERENCES users(id),
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY (group_id, user_id)
);
