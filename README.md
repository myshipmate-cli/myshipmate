# Shipmate

**The smart deployer for developers**

Shipmate is a CLI tool that detects your project type and deploys it to the right platform automatically. One command, zero configuration.

## Installation

```bash
curl -fsSL myshipmate.cc/install | sh
```

Or install with Go:

```bash
go install github.com/myshipmate-cli/myshipmate@latest
```

## Quick Start

```bash
cd your-project
shipmate
```

That's it. Shipmate will:
1. Detect your project type (Next.js, Go, Node.js, Python, etc.)
2. Recommend the best deployment platform
3. Handle authentication and deployment
4. Give you a live URL

## Supported Platforms

- **Vercel** - Best for Next.js, React, and static sites
- **Railway** - Great for Go, Node.js, Python, and Docker
- **Render** - Simple deployment for backends and full-stack apps
- **Netlify** - Excellent for static sites and JAMstack
- **Fly.io** - Global edge deployment for Docker containers

## Supported Project Types

- Next.js
- React / Vue / Svelte
- Go
- Node.js / Express
- Python (Django, Flask, FastAPI)
- Ruby on Rails
- Static HTML/CSS/JS
- Docker projects
- Monorepos (Turborepo, Nx, pnpm workspaces)

## Example Usage

```bash
$ cd my-nextjs-app
$ shipmate

🚀 Shipmate - Smart Deployer

🔍 Scanning project...
   ✓ Found next.config.js
   ✓ Detected Next.js 14 with App Router
   ✓ Found Tailwind CSS

Recommended platforms:
  ❯ Vercel (built for Next.js, best performance)
    Netlify (also great for Next.js)
    Railway (if you need a database too)

🔐 Authenticating with Vercel...
   ✓ Authorized

🚀 Deploying...
   ✓ Deployed to https://my-nextjs-app.vercel.app
```

## Features

- ✅ **Smart Detection** - Automatically detects project type and structure
- ✅ **Platform Recommendations** - Suggests the best platforms for your project
- ✅ **Environment Variables** - Syncs your `.env` file to the deployment platform
- ✅ **Monorepo Support** - Works with Turborepo, Nx, and pnpm workspaces
- ✅ **One Command** - No config files, no manual setup
- ✅ **Platform Switching** - Easily deploy to different platforms

## Project Structure

```
shipmate/
├── cli/              # Go CLI application
│   ├── main.go       # CLI entry point
│   ├── detector.go   # Project type detection
│   └── platforms.go  # Platform configuration
├── web/              # Next.js landing page and dashboard
│   ├── src/
│   │   └── app/
│   │       ├── page.tsx
│   │       ├── layout.tsx
│   │       └── globals.css
│   └── package.json
└── README.md
```

## Development

### CLI (Go)

```bash
cd cli
go mod tidy
go run main.go ship
```

### Web (Next.js)

```bash
cd web
pnpm install
pnpm dev
```

Then open http://localhost:3000

## Roadmap

### MVP (Current)
- [x] Project structure
- [x] Basic CLI skeleton
- [x] Landing page
- [ ] Project detection logic
- [ ] Vercel deployment integration
- [ ] Railway deployment integration
- [ ] Environment variable sync

### V2 (Future)
- [ ] Monorepo support
- [ ] Multi-project detection
- [ ] Netlify, Render, Fly.io integration
- [ ] Deployment history
- [ ] Better error handling

### V3 (Paid Features)
- [ ] Web dashboard
- [ ] Team collaboration
- [ ] Git Ship Cloud (deploy through our accounts)
- [ ] Rollback and environment promotion

## About the Database Question

**Do we need a database?**

For MVP: **No.**
- The CLI stores tokens locally in `~/.shipmate/config.json`
- The landing page is static
- No user accounts needed yet

For V3 (Paid Dashboard): **Yes.**
- User accounts for the dashboard
- Deployment history tracking
- Team features
- Git Ship Cloud (deploying through our accounts)

We'll add a database (Postgres) when we build the paid dashboard features.

## Contributing

This is currently a solo project, but contributions are welcome! Please open an issue or PR.

## License

MIT

## Links

- Website: https://myshipmate.cc
- GitHub: https://github.com/myshipmate-cli/myshipmate
- Twitter: @shipmate_dev

---

Built with ❤️ for developers who just want to ship.
