CREATE TABLE
    IF NOT EXISTS orders (
        order_uid VARCHAR PRIMARY KEY,
        track_number VARCHAR NOT NULL,
        entry VARCHAR NOT NULL,
        locale VARCHAR NOT NULL,
        internal_signature VARCHAR,
        customer_id VARCHAR NOT NULL,
        delivery_service VARCHAR NOT NULL,
        shardkey VARCHAR NOT NULL,
        sm_id INTEGER NOT NULL,
        date_created VARCHAR NOT NULL,
        oof_shard VARCHAR NOT NULL
    );

CREATE TABLE
    IF NOT EXISTS delivery (
        order_uid VARCHAR PRIMARY KEY REFERENCES orders (order_uid),
        name VARCHAR NOT NULL,
        phone VARCHAR NOT NULL,
        zip VARCHAR NOT NULL,
        city VARCHAR NOT NULL,
        address VARCHAR NOT NULL,
        region VARCHAR NOT NULL,
        email VARCHAR NOT NULL
    );

CREATE TABLE
    IF NOT EXISTS payment (
        order_uid VARCHAR PRIMARY KEY REFERENCES orders (order_uid),
        transaction VARCHAR NOT NULL,
        request_id VARCHAR,
        currency VARCHAR NOT NULL,
        provider VARCHAR NOT NULL,
        amount INTEGER NOT NULL,
        payment_dt INTEGER NOT NULL,
        bank VARCHAR NOT NULL,
        delivery_cost INTEGER NOT NULL,
        goods_total INTEGER NOT NULL,
        custom_fee INTEGER NOT NULL
    );

CREATE TABLE
    IF NOT EXISTS items (
        order_uid VARCHAR REFERENCES orders (order_uid),
        chrt_id INTEGER NOT NULL,
        track_number VARCHAR NOT NULL,
        price INTEGER NOT NULL,
        rid VARCHAR NOT NULL,
        name VARCHAR NOT NULL,
        sale INTEGER NOT NULL,
        size VARCHAR NOT NULL,
        total_price INTEGER NOT NULL,
        nm_id INTEGER NOT NULL,
        brand VARCHAR NOT NULL,
        status INTEGER NOT NULL,
        PRIMARY KEY (order_uid, chrt_id)
    );