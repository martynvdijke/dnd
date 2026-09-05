package migrations

var migration016SQL = `
CREATE TABLE IF NOT EXISTS shops (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    campaign_id INTEGER REFERENCES campaigns(id) ON DELETE SET NULL,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    markup_percent REAL NOT NULL DEFAULT 100,
    markup_buy_percent REAL NOT NULL DEFAULT 50,
    location_id INTEGER REFERENCES locations(id) ON DELETE SET NULL,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS shop_items (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    shop_id INTEGER NOT NULL REFERENCES shops(id) ON DELETE CASCADE,
    item_name TEXT NOT NULL,
    category TEXT NOT NULL DEFAULT 'gear',
    price_gp REAL NOT NULL DEFAULT 0,
    quantity_available INTEGER NOT NULL DEFAULT -1,
    description TEXT NOT NULL DEFAULT '',
    is_magical INTEGER NOT NULL DEFAULT 0,
    attunement_required INTEGER NOT NULL DEFAULT 0,
    notes TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS shop_transactions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    shop_id INTEGER NOT NULL REFERENCES shops(id) ON DELETE CASCADE,
    character_id INTEGER NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    item_name TEXT NOT NULL,
    quantity INTEGER NOT NULL DEFAULT 1,
    price_gp REAL NOT NULL DEFAULT 0,
    transaction_type TEXT NOT NULL CHECK(transaction_type IN ('buy','sell')),
    timestamp TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS campaign_wiki_pages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    campaign_id INTEGER NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    parent_id INTEGER REFERENCES campaign_wiki_pages(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    content TEXT NOT NULL DEFAULT '',
    visibility TEXT NOT NULL DEFAULT 'public' CHECK(visibility IN ('public','dm-only')),
    tags TEXT NOT NULL DEFAULT '[]',
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_shops_campaign ON shops(campaign_id);
CREATE INDEX IF NOT EXISTS idx_shop_items_shop ON shop_items(shop_id);
CREATE INDEX IF NOT EXISTS idx_shop_transactions_shop ON shop_transactions(shop_id);
CREATE INDEX IF NOT EXISTS idx_shop_transactions_char ON shop_transactions(character_id);
CREATE INDEX IF NOT EXISTS idx_wiki_campaign ON campaign_wiki_pages(campaign_id);
CREATE INDEX IF NOT EXISTS idx_wiki_parent ON campaign_wiki_pages(parent_id);
`
