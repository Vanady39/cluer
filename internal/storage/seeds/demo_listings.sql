INSERT INTO listings (title, description, price, image_url)
SELECT seed.title, seed.description, seed.price, seed.image_url
FROM (
    VALUES
        ('iPhone 15 Pro', 'Телефон в отличном состоянии, без царапин', 95000, 'https://example.com/iphone.png'),
        ('Игровое кресло', 'Использовалось полгода, удобная поддержка спины', 15000, 'https://example.com/chair.png'),
        ('Samsung Galaxy S24', 'Новый телефон в заводской упаковке', 72000, 'https://example.com/samsung.png')
) AS seed(title, description, price, image_url)
WHERE NOT EXISTS (SELECT 1 FROM listings);
