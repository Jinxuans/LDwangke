# Legacy Surface

## Decision

`internal/legacy/*` stays. It is a compatibility layer, not a refactor target for broad removal.

The current rule is:

- Keep the legacy routes stable.
- Do not add new product logic into `internal/legacy/*`.
- Only change legacy behavior for compatibility fixes, security fixes, or extraction work with a concrete replacement.

## Current legacy entrypoints

### `internal/legacy/openapi`

Compatibility routes:

- `GET|POST /api.php`
- `GET|POST /api/index.php`

Protected OpenAPI routes:

- `GET|POST /api/v1/open/classlist`
- `GET|POST /api/v1/open/query`
- `GET|POST /api/v1/open/order`
- `GET|POST /api/v1/open/orderlist`
- `GET /api/v1/open/balance`
- `GET|POST /api/v1/open/chadan`
- `POST /api/v1/open/bindpushuid`
- `POST /api/v1/open/bindpushemail`
- `POST /api/v1/open/bindshowdocpush`

Status:

- Keep.
- These are clearly external protocol surfaces.
- Internally they still call `service` and raw SQL in places, but that is acceptable as a compatibility host.

### `internal/legacy/php`

Routes:

- `GET|POST /php-api/*path`
- `POST /internal/php-bridge/money`
- `GET /internal/php-bridge/user`
- `POST /internal/php-bridge/order`
- `GET /api/v1/php-bridge/auth-url`

Status:

- Keep.
- `php-api` proxy and bridge endpoints are runtime integration points with the PHP side.
- They should be treated as infrastructure compatibility, not business-domain APIs.

### `internal/legacy/module`

Routes:

- `GET /api/v1/module/:app_id/frame-url`
- `GET|POST /api/v1/module/:app_id`

Status:

- Keep for now.
- This is the weakest long-term surface because it proxies module traffic into PHP by `app_id`.
- It is a candidate for future shrinkage only after module-by-module replacement exists.

## Deprecation posture

Current stance by surface:

- `legacy/openapi`: no deprecation yet
- `legacy/php`: no deprecation yet
- `legacy/module`: candidate for future deprecation, but not before explicit replacement planning

## Allowed change types

Allowed:

- Security hardening
- Bug fixes
- Runtime config fixes
- Documentation
- Tests
- Narrow extraction of reusable logic out of `service` or `legacy`

Not allowed by default:

- New domain logic added directly into `legacy/*`
- New business routes added under old compatibility prefixes
- Silent route shape changes
