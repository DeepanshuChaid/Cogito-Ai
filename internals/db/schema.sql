CREATE TABLE IF NOT EXISTS sessions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT UNIQUE NOT NULL,
    project TEXT NOT NULL,
    started_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    status TEXT DEFAULT 'active',
    user_prompt TEXT
)

CREATE TABLE IF NOT EXISTS observations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT NOT NULL,
    project TEXT NOT NULL,
    observation_type TEXT,
    raw_text TEXT NOT NULL,
    compressed_text TEXT,
    files_touched TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FORGEIGN KEY(session_id) REFERENCES sessions(sessions_id))
)

CREATE TABLE IF NOT EXISTS summaries (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT UNIQUE NOT NULL,
    project TEXT NOT NULL,
    summary_text TEXT,
    learned_facts TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(session_id) REFERENCES sessions(session_id)
)
