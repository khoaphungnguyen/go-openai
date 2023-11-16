-- User Table
CREATE TABLE "user" (
  id UUID PRIMARY KEY,
  full_name VARCHAR(100) NOT NULL,
  email VARCHAR(50) UNIQUE NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  salt VARCHAR(255),
  role VARCHAR(10) DEFAULT 'user',
  email_verified BOOLEAN DEFAULT FALSE,
  last_login TIMESTAMPTZ,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW(),
  deleted_at TIMESTAMPTZ DEFAULT NULL
);

-- Updated Chat Message Table to include user questions and OpenAI responses
CREATE TABLE "chat_message" (
  id SERIAL PRIMARY KEY,
  user_id UUID REFERENCES "user"(id) ON DELETE SET NULL, -- Set to null if user is deleted
  role VARCHAR(50) NOT NULL CHECK (role IN ('user', 'assistant')), -- Role of the message sender
  model VARCHAR(255), -- Model of the AI (e.g., 'gpt-3.5-turbo')
  content TEXT NOT NULL, -- Content of the message
  created_at TIMESTAMPTZ DEFAULT NOW(), -- Timestamp of the message creation
  updated_at TIMESTAMPTZ DEFAULT NOW() -- Timestamp of the last update
);

