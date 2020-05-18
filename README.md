# ghrp

URL format: /`owner`/`repo`/`query`

The first asset whose name contains `query` is returned.

Example URL: [https://release.khlieng.com/khlieng/dispatch/linux_x64](https://release.khlieng.com/khlieng/dispatch/linux_x64)

## Usage

```bash
go get github.com/khlieng/ghrp
GITHUB_TOKEN=<token> ghrp
```

#### Environment variables

- `GITHUB_TOKEN` A GitHub token, no scopes needed, this is required
- `PORT` Port to listen on, defaults to 3001
