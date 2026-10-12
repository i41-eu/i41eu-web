# i41eu-web

The page at [i41.eu](https://i41.eu): what the i41 home platform is, with links to its parts
([s.i41.eu](https://s.i41.eu), [m.i41.eu](https://m.i41.eu), the apps on q41.eu) and to the
code on [GitHub](https://github.com/i41-eu). It also serves [s.i41.eu](https://s.i41.eu), the
Stackchan robot's page: what the robot does, a page per app (s.i41.eu/pet, /focus, /raw, /dev)
and the pictures (`s/`). The robot in 3D at s.i41.eu/s3d/ is
[i41eu-s3d](https://github.com/i41-eu/i41eu-s3d).

One Go binary with the pages embedded, chosen by the request's host: `s.html`, `app.html`
around `apps/<app>.html` and `s/*` for s.i41.eu, `index.html` for any other host; `/healthz`
for probes. The robot pictures in `s/` come from
[i41eu-assets](https://github.com/i41-eu/i41eu-assets) (`robot3d-distribute`).

```bash
go run . -listen :8080
go test ./...
```

A tag `v*` builds the image `ghcr.io/i41-eu/i41eu-web:<tag>` (`.github/workflows/release.yml`).

License: Apache-2.0.
