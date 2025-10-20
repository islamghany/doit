INSERT INTO todos (user_id, title, description, status)
VALUES (
        '550e8400-e29b-41d4-a716-446655440000',
        'Welcome Todo',
        'This is your first todo',
        'pending'
    ) ON CONFLICT DO NOTHING;