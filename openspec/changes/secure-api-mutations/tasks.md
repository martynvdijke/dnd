## 1. Route and data audit

- [ ] 1.1 Inventory all mutation groups and existing session/admin/ownership checks.
- [ ] 1.2 Add reversible user-token migration with indexes and revocation metadata.

## 2. Authentication

- [ ] 2.1 Implement bearer parsing, hash verification, expiry, revocation, and owner matching.
- [ ] 2.2 Keep the three TRMNL reads public and protect all writes.
- [ ] 2.3 Add user token create/list/revoke/rotate flows with one-time secrets.

## 3. Verification

- [ ] 3.1 Test public polling, token-only rejection, malformed/expired/revoked tokens, and ownership.
- [ ] 3.2 Test admin boundaries and ensure secrets never appear in logs or later responses.
- [ ] 3.3 Update plugin/client documentation and run the full Villum CI suite.
