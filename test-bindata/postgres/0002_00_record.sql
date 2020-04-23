-- +migrate Up
INSERT INTO public.markets(symbol, base_currency, quote_currency)
VALUES ('BTCUSDT', 'BTC', 'USDT'),
       ('BCHUSDT', 'BCH', 'USDT'),
       ('ETHUSDT', 'ETH', 'USDT'),
       ('LTCUSDT', 'LTC', 'USDT'),
       ('XRPUSDT', 'XRP', 'USDT'),
       ('BCHBTC', 'BCH', 'BTC'),
       ('ETHBTC', 'ETH', 'BTC'),
       ('LTCBTC', 'LTC', 'BTC'),
       ('XRPBTC', 'XRP', 'BTC'),
       ('ETHBCH', 'ETH', 'BCH'),
       ('LTCBCH', 'LTC', 'BCH'),
       ('XRPBCH', 'XRP', 'BCH'),
       ('LTCETH', 'LTC', 'ETH'),
       ('XRPETH', 'XRP', 'ETH'),
       ('XRPLTC', 'XRP', 'LTC');

-- +migrate Down
DELETE FROM markets;
