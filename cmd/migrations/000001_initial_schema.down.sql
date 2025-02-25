-- Drop all tables in reverse order of creation to avoid foreign key constraints
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS post_tags;
DROP TABLE IF EXISTS tags;
DROP TABLE IF EXISTS login_providers;
DROP TABLE IF EXISTS comment_votes;
DROP TABLE IF EXISTS post_votes;
DROP TABLE IF EXISTS post_comments_metadata;
DROP TABLE IF EXISTS comments;
DROP TABLE IF EXISTS posts;
DROP TABLE IF EXISTS email_confirmations;
DROP TABLE IF EXISTS users;

-- Drop the UUID-OSSP extension
DROP EXTENSION IF EXISTS "uuid-ossp";