CREATE TABLE user_link_requests (
	id uuid PRIMARY KEY NOT NULL,
	user_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
	agent_id uuid NOT NULL REFERENCES workspace_agents (id) ON DELETE CASCADE,
	created_at timestamptz NOT NULL,
	updated_at timestamptz NOT NULL,
	expires_at timestamptz NOT NULL,
	provider text[] NOT NULL,
	login_user text NOT NULL,
	login_url text NOT NULL,
	resolved boolean NOT NULL DEFAULT false
);

COMMENT ON TABLE user_link_requests IS 'User link requests are used to prompt for authentication (OAuth/OIDC) in the browser, e.g. by git askpass inside workspaces.';
COMMENT ON COLUMN user_link_requests.id IS 'The unique ID of the request.';
COMMENT ON COLUMN user_link_requests.user_id IS 'The ID of the user this request belongs to.';
COMMENT ON COLUMN user_link_requests.agent_id IS 'The ID of the agent that created this request.';
COMMENT ON COLUMN user_link_requests.created_at IS 'The time the request was created.';
COMMENT ON COLUMN user_link_requests.updated_at IS 'The time the request was updated.';
COMMENT ON COLUMN user_link_requests.expires_at IS 'The time at which this request expires, if unresolved.';
COMMENT ON COLUMN user_link_requests.provider IS 'The auth provider that was matched for this request, can have multiple values in case of conflict.';
COMMENT ON COLUMN user_link_requests.login_user IS 'The requested user for the login request, often empty.';
COMMENT ON COLUMN user_link_requests.login_url IS 'The URL this login was requested for.';
COMMENT ON COLUMN user_link_requests.resolved IS 'Resolved is set to true when the login is successful.';

ALTER TABLE user_links
	ADD COLUMN login_user text NOT NULL DEFAULT '';

COMMENT ON COLUMN user_links.login_user IS 'The login user this link refers to (can be different from Coder user).';

ALTER TABLE user_links
	ADD COLUMN default_login_user boolean NOT NULL DEFAULT FALSE;

COMMENT ON COLUMN user_links.default_login_user IS 'Default login user indicates this link will be used when no specific user is requested.';

ALTER TABLE user_links
	ADD COLUMN login_url text NOT NULL DEFAULT '';

COMMENT ON COLUMN user_links.login_url IS 'The login url links this entry to a specific provider.';

ALTER TABLE user_links
	ADD COLUMN scopes text[] NOT NULL DEFAULT '{}';

COMMENT ON COLUMN user_links.scopes IS 'The current scopes available for this user link, can be used to verify if reauth is required to receive additional permissions.';

ALTER TABLE user_links DROP CONSTRAINT user_links_pkey;
ALTER TABLE user_links ADD CONSTRAINT user_links_pkey PRIMARY KEY (user_id, login_type, login_user, login_url);

ALTER TABLE api_keys
	ADD COLUMN login_url text NOT NULL DEFAULT '';
