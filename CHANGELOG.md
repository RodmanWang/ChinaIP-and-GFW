# Changelog

## 1.1.0 - 2026-11-17

- Upgrade the module baseline to Go 1.23 and update MMDB writer and GeoIP2 reader dependencies.
- Reject malformed CIDR entries instead of silently omitting them from generated databases.
- Increase scanner capacity and report scanner errors with line context.
- Write MMDB output to a temporary file and atomically replace the destination only after a successful write.
- Close files explicitly and improve CLI error context.
- Make the build script reproducible with `-trimpath`, stripped symbols, and guaranteed `dist/` creation.
- Make the verifier accept a database path and optional IP arguments, and handle invalid addresses without panicking.
- Add CIDR parsing regression tests.
