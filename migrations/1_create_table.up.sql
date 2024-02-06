-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Users Table
CREATE TABLE IF NOT EXISTS users (
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
CREATE TABLE IF NOT EXISTS chat_thread (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  title VARCHAR(255),
  model VARCHAR(255),
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Chat Message Table
CREATE TABLE IF NOT EXISTS chat_message (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  thread_id UUID REFERENCES chat_thread(id) ON DELETE CASCADE,
  user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  role VARCHAR(50) NOT NULL CHECK (role IN ('user', 'assistant')),
  content TEXT NOT NULL,
  created_at TIMESTAMPTZ DEFAULT NOW()
);

-- OpenAI Transacton Table
CREATE TABLE IF NOT EXISTS openai_transaction (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  thread_id UUID REFERENCES chat_thread(id) ON DELETE CASCADE,
  message_id UUID REFERENCES chat_message(id) ON DELETE SET NULL,
  model VARCHAR(255),
  role VARCHAR(50) NOT NULL CHECK (role IN ('user', 'assistant')),
  message_length INT,  -- Tracks the length of the user's message
  process_time TIMESTAMPTZ DEFAULT NOW()
);


-- OpenAI Transacton Table
CREATE TABLE IF NOT EXISTS openai_transaction (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  thread_id UUID REFERENCES chat_thread(id) ON DELETE CASCADE,
  message_id UUID REFERENCES chat_message(id) ON DELETE SET NULL,
  model VARCHAR(255),
  role VARCHAR(50) NOT NULL CHECK (role IN ('user', 'assistant')),
  message_length INT,  -- Tracks the length of the user's message
  process_time TIMESTAMPTZ DEFAULT NOW()
);

-- Note Table
CREATE TABLE IF NOT EXISTS notes (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  title VARCHAR(255),
  problem TEXT,
  approach TEXT,
  solution TEXT,
  code TEXT,
  level VARCHAR(50) NOT NULL CHECK (level IN ('Easy', 'Medium', 'Hard')),
  type TEXT,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create or replace the trigger function to update the chat_thread
CREATE OR REPLACE FUNCTION update_thread_on_new_message()
RETURNS TRIGGER AS $$
BEGIN
    -- Update the updated_at field and set the title to the first 40 characters of the new message content
    UPDATE chat_thread
    SET 
        updated_at = NOW(),
        title = LEFT(NEW.content, 40)  -- Truncate content to 40 characters
    WHERE id = NEW.thread_id;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- The trigger itself remains the same
CREATE TRIGGER trigger_update_thread_on_new_message
AFTER INSERT ON chat_message
FOR EACH ROW
EXECUTE FUNCTION update_thread_on_new_message();

