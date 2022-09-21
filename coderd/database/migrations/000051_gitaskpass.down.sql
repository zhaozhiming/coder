ALTER TABLE user_links DROP CONSTRAINT user_links_pkey;
ALTER TABLE user_links ADD CONSTRAINT user_links_pkey PRIMARY KEY (user_id, login_type);

ALTER TABLE user_links
    DROP COLUMN login_origin;

ALTER TABLE api_keys
    DROP COLUMN login_origin;
