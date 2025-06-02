INSERT INTO apps (id, name, secret)
VALUES (1, 'test-app', 'secret')
ON CONFLICT DO NOTHING;