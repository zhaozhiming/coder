-- name: InsertUserLinkRequest :one
INSERT INTO
	user_link_requests (
		id,
		agent_id,
		user_id,
		created_at,
		updated_at,
		expires_at,
		provider,
		login_user,
		login_url,
		resolved
	)
VALUES
	($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING *;

-- name: GetUserLinkRequestsByUserID :many
SELECT
	*
FROM
	user_link_requests
WHERE
	user_id = $1
	AND CASE
		WHEN @status = 'expired' THEN
			resolved = FALSE AND expires_at < NOW()
		WHEN @status = 'pending' THEN
			resolved = FALSE AND expires_at >= NOW()
		WHEN @status = 'resolved' THEN
			resolved = TRUE
		ELSE
			TRUE
	END;

-- name: GetUserLinkRequestByIDAndUserID :one
SELECT
	*
FROM
	user_link_requests
WHERE
	id = $1
	AND user_id = $2;

-- name: GetUserLinkRequestByIDAndAgentID :one
SELECT
	*
FROM
	user_link_requests
WHERE
	id = $1
	AND agent_id = $2;

-- name: UpdateUserLinkRequestByID :exec
UPDATE
	user_link_requests
SET
	updated_at = $1,
	resolved = $2
WHERE
	id = $3
RETURNING *;
