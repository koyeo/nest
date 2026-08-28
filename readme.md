# ship

pnpm monorepo.

| Package | Path | Description |
|:--|:--|:--|
| `nest` | `apps/nest` | Go CLI (local-first CI/CD). See `apps/nest/readme.md`. |
| `@kozilla/ship` | `apps/ship` | TypeScript port of nest's deploy features (`ship.yaml`, no UI mode). |
| `nest-docs` | `docs` | VitePress documentation site. |

```bash
pnpm install
pnpm nest:build   # -> apps/nest/build/nest
pnpm nest:test
pnpm ship:build   # -> apps/ship/dist/main.js
pnpm ship:test
pnpm docs:dev
```
