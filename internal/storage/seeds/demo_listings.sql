INSERT INTO listings (id, title, description, price, image_url) OVERRIDING SYSTEM VALUE
VALUES
    (1, 'iPhone 15 Pro', 'Телефон в отличном состоянии, без царапин', 95000, 'https://example.com/iphone.png'),
    (2, 'Игровое кресло', 'Использовалось полгода, удобная поддержка спины', 15000, 'https://example.com/chair.png'),
    (3, 'Samsung Galaxy S24', 'Новый телефон в заводской упаковке', 72000, 'https://example.com/samsung.png')
ON CONFLICT (id) DO NOTHING;

SELECT setval(
    pg_get_serial_sequence('listings', 'id'),
    (SELECT MAX(id) FROM listings),
    true
);
