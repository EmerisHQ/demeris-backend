-- name: GetChain :one
SELECT * FROM chains WHERE chain_name='?' limit 1;
