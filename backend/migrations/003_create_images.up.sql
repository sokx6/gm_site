CREATE TABLE images (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    album_id INTEGER REFERENCES albums(id) ON DELETE SET NULL,
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    tags TEXT NOT NULL DEFAULT '[]',
    lsky_url TEXT NOT NULL,
    thumbnail_url TEXT NOT NULL DEFAULT '',
    uploaded_by INTEGER NOT NULL REFERENCES users(id),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_images_album_id ON images(album_id);
CREATE INDEX idx_images_uploaded_by ON images(uploaded_by);
CREATE INDEX idx_images_created_at ON images(created_at);
