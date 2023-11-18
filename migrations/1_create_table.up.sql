-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Users Table
CREATE TABLE users (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
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

-- Chat Thread Table
CREATE TABLE chat_thread (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  title VARCHAR(255),
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- OpenAI Transacton Table
CREATE TABLE openai_transaction (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  thread_id UUID REFERENCES chat_thread(id) ON DELETE CASCADE,
  message_id UUID REFERENCES chat_message(id) ON DELETE SET NULL,
  model VARCHAR(255),
  role VARCHAR(50) NOT NULL CHECK (role IN ('user', 'assistant')),
  message_length INT,  -- Tracks the length of the user's message
  process_time TIMESTAMPTZ DEFAULT NOW()
);
