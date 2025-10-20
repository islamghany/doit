INSERT INTO users (id, email, username, password_hash)
VALUES (
        '550e8400-e29b-41d4-a716-446655440000',
        'admin@example.com',
        'admin',
        '$2a$10$YourHashHere'
    ),
    (
        '550e8400-e29b-41d4-a716-446655440001',
        'user@example.com',
        'user',
        '$2a$10$YourHashHere'
    ) ON CONFLICT (email) DO NOTHING;