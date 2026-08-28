# ship

pnpm monorepo.

| Package | Path | Description |
|:--|:--|:--|
| `nest` | `apps/nest` | Go CLI (local-first CI/CD). See `apps/nest/readme.md`. |
| `nest-docs` | `docs` | VitePress documentation site. |

```bash
pnpm install
pnpm nest:build   # -> apps/nest/build/nest
pnpm nest:test
pnpm docs:dev
```
