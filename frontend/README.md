# Feature Flag Backoffice

React backoffice for the API described in `../docs/openapi.yaml`.

```sh
npm install
npm run dev
```

The Vite server proxies `/api` and `/healthz` to `http://localhost:8080`. Start the Go API and its dependencies first. For a separately hosted build, set `VITE_API_BASE_URL` (for example, `https://api.example.test`) before running `npm run build`.

Use `npm run build` for type-checking and a production bundle, and `npm test` for the component and API-client tests.
