ALTER TABLE user_links
    ADD COLUMN login_origin text NOT NULL DEFAULT '';

ALTER TABLE user_links DROP CONSTRAINT user_links_pkey;
ALTER TABLE user_links ADD CONSTRAINT user_links_pkey PRIMARY KEY (user_id, login_type, login_origin);

ALTER TABLE api_keys
    ADD COLUMN login_origin text NOT NULL DEFAULT '';
