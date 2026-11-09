CREATE TABLE IF NOT EXISTS default_message_style (
  singleton boolean PRIMARY KEY DEFAULT TRUE CHECK (singleton = TRUE),
  font_family varchar(64) NOT NULL,
  font_size smallint NOT NULL CHECK (font_size BETWEEN 10 AND 28),
  color char(7) NOT NULL CHECK (color ~ '^#[0-9A-Fa-f]{6}$'),
  bold boolean NOT NULL DEFAULT FALSE,
  italic boolean NOT NULL DEFAULT FALSE,
  underline boolean NOT NULL DEFAULT FALSE,
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS room_message_styles (
  room_id varchar(128) PRIMARY KEY,
  font_family varchar(64) NOT NULL,
  font_size smallint NOT NULL CHECK (font_size BETWEEN 10 AND 28),
  color char(7) NOT NULL CHECK (color ~ '^#[0-9A-Fa-f]{6}$'),
  bold boolean NOT NULL DEFAULT FALSE,
  italic boolean NOT NULL DEFAULT FALSE,
  underline boolean NOT NULL DEFAULT FALSE,
  updated_at timestamptz NOT NULL DEFAULT now()
);

INSERT INTO default_message_style(singleton,font_family,font_size,color,bold,italic,underline)
VALUES(TRUE,'Arial',14,'#FFFFFF',FALSE,FALSE,FALSE)
ON CONFLICT(singleton) DO NOTHING;
