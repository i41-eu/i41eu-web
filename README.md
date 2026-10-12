# i41eu-web

The page at [i41.eu](https://i41.eu): what the i41 home platform is, with links to its parts
([s.i41.eu](https://s.i41.eu), [m.i41.eu](https://m.i41.eu), the apps on q41.eu) and to the
code on [GitHub](https://github.com/i41-eu).

One Go binary with the page embedded (`index.html`); `/healthz` for probes.

```bash
go run . -listen :8080
go test ./...
```

A tag `v*` builds the image `ghcr.io/i41-eu/i41eu-web:<tag>` (`.github/workflows/release.yml`).

License: Apache-2.0.
