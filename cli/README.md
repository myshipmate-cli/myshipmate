# Shipmate CLI

**The smart deployer for developers**

Shipmate is a CLI tool that detects your project type, generates optimized Dockerfiles, and deploys to the right platform automatically.

## Installation

### From Source (Go required)

```bash
go install github.com/myshipmate-cli/myshipmate@latest
```

### Binary Download (Coming Soon)

```bash
curl -fsSL myshipmate.cc/install | sh
```

## Quick Start

```bash
cd your-project
shipmate
```

That's it. Shipmate will:
1. Detect your project type (framework, language, structure)
2. Generate an optimized Dockerfile (if you don't have one)
3. Recommend the best deployment platforms
4. Deploy with one command

## Commands

### `shipmate` or `shipmate ship`

The main command. Scans your project, generates Dockerfile if needed, and prepares for deployment.

```bash
$ shipmate

🚀 Shipmate - The Smart Deployer for Developers

🔍 Scanning project...
   ✓ Project: my-nextjs-app
   ✓ Type: nextjs
   ✓ Environment file detected

📦 Generating Dockerfile...
   ✓ Dockerfile created
   ✓ .dockerignore created

ℹ️  No local Docker needed - deployment platforms will build your Dockerfile

🎯 Recommended platforms:
  ❯ Vercel
    Netlify
    Railway

📝 Next steps:
   1. Review the generated Dockerfile
   2. Run: shipmate deploy <platform>
   3. Authenticate with your chosen platform
   4. Shipmate will push your code and deploy

💡 Note: You don't need Docker installed locally.
   The deployment platform builds your Dockerfile in their cloud.
```

### `shipmate init`

Initialize Shipmate in your project without deploying. Generates Dockerfile and .dockerignore only.

```bash
$ shipmate init

🔧 Initializing Shipmate...

   ✓ Detected: go project
   ✓ Dockerfile created
   ✓ .dockerignore created

✓ Shipmate initialized!

Next: Run 'shipmate ship' to deploy your project
```

### `shipmate deploy [platform]`

Deploy to a specific platform directly.

```bash
$ shipmate deploy vercel
🚀 Deploying to vercel...
   (Coming soon)
```

## Supported Languages & Frameworks

### JavaScript/TypeScript
- **Next.js** - Optimized multi-stage Dockerfile with standalone output
- **React** - Build + nginx serving
- **Vue** - Build + nginx serving
- **Angular** - Build + nginx serving
- **Svelte / SvelteKit** - Build + node for SSR
- **Nuxt** - Build + node for SSR
- **Remix** - Build + node serving
- **Astro** - Build + nginx or node
- **TanStack** - Build + nginx serving
- **Gatsby** - Build + nginx serving
- **SolidJS** - Build + nginx serving
- **Qwik** - Build + nginx or node

### Backend Languages
- **Go** - Multi-stage build with Alpine runtime
- **Node.js** - Node 20 Alpine
- **Python** - Python 3.11 slim
  - **Django** - With gunicorn + static file collection
  - **Flask** - With gunicorn
  - **FastAPI** - With uvicorn
- **Ruby** - Ruby 3.2 slim
  - **Rails** - With asset precompilation
  - **Sinatra** - Lightweight setup
- **Java / Kotlin** - Maven build + JRE runtime
  - **Spring Boot** - Optimized for Spring
- **Rust** - Multi-stage build with release profile
- **C# / .NET** - .NET 8 SDK build + ASP.NET runtime
- **PHP** - PHP 8.2 with Apache
  - **Laravel** - With Composer + PHP-FPM
  - **Symfony** - With Composer + PHP-FPM
- **Elixir** - Mix release + Alpine runtime
- **Dart / Flutter** - Flutter web build + nginx

### Static Sites
- Plain HTML/CSS/JS - nginx serving

## How It Works

### 1. Project Detection

Shipmate scans your project directory and looks for key files:

| File | Detected As |
|------|-------------|
| `go.mod` | Go |
| `Cargo.toml` | Rust |
| `pom.xml`, `build.gradle` | Java/Kotlin |
| `Gemfile` | Ruby |
| `requirements.txt`, `Pipfile`, `pyproject.toml` | Python |
| `package.json` + `next.config.js` | Next.js |
| `package.json` + `vue` dependency | Vue |
| `package.json` + `@angular/core` | Angular |
| `package.json` + `svelte` | Svelte |
| `composer.json` + `laravel/framework` | Laravel |
| `index.html` (no build step) | Static |

### 2. Dockerfile Generation

Shipmate generates **optimized Dockerfiles** specific to your project type:

**Example: Next.js**
```dockerfile
# Multi-stage build for minimal production image
FROM node:20-alpine AS builder
# ... build stage ...

FROM node:20-alpine AS runner
# ... production stage with only necessary files ...
```

**Example: Go**
```dockerfile
# Multi-stage build with Alpine runtime
FROM golang:1.21-alpine AS builder
# ... build binary ...

FROM alpine:latest
# ... copy only the binary ...
```

**Example: Python/Django**
```dockerfile
FROM python:3.11-slim
# ... install dependencies ...
# ... collect static files ...
# ... run with gunicorn ...
```

### 3. Platform Deployment

Shipmate recommends platforms based on your project type:

| Project Type | Recommended Platforms |
|--------------|----------------------|
| Next.js, React, Vue, Angular | Vercel, Netlify, Railway |
| Go, Node.js, Python, Java | Railway, Render, Fly.io |
| Static sites | Netlify, Vercel, Cloudflare Pages |
| Docker projects | Railway, Fly.io, Render |

## Do I Need Docker Installed?

**No!** This is a key feature of Shipmate.

- Shipmate generates the Dockerfile
- You push your source code to the deployment platform
- The platform (Railway, Fly.io, Render, etc.) builds the Docker image in their cloud
- Your app goes live

This means:
- ✅ No Docker installation required
- ✅ No local build time
- ✅ No disk space used for images
- ✅ Works on any machine with the CLI

## Environment Variables

Shipmate automatically detects `.env` files and syncs them to your deployment platform:

```bash
$ shipmate

🔍 Scanning project...
   ✓ Environment file detected (.env)

🚀 Deploying...
   ✓ Synced 5 environment variables to Railway
   ✓ DATABASE_URL
   ✓ API_KEY
   ✓ JWT_SECRET
   ✓ PORT
   ✓ NODE_ENV
```

## Monorepo Support

Shipmate detects monorepo structures (Turborepo, Nx, pnpm workspaces) and handles them intelligently:

```bash
$ shipmate

🔍 Scanning project...
   ✓ Detected monorepo structure (Turborepo)
   
Packages:
  📁 apps/web (Next.js)
  📁 apps/api (Node.js)
  📁 packages/ui (shared)

What do you want to deploy?
  ❯ apps/web
    apps/api
    Both
```

## Project Structure

```
shipmate/
├── cli/
│   ├── main.go                 # CLI entry point (cobra)
│   ├── detector.go             # Project type detection
│   ├── platforms.go            # Platform recommendations
│   ├── dockerfile_generator.go # Dockerfile generation
│   └── go.mod                  # Go dependencies
├── web/
│   ├── src/app/
│   │   ├── page.tsx           # Landing page
│   │   ├── layout.tsx         # Root layout
│   │   └── globals.css        # Styles
│   └── package.json           # Dependencies
└── README.md
```

## Development

### Build from source

```bash
cd cli
go mod tidy
go build -o shipmate .
./shipmate --help
```

### Test in a project

```bash
# In any project directory
~/shipmate/cli/shipmate
```

## Roadmap

### Current (MVP)
- ✅ Project detection (30+ languages/frameworks)
- ✅ Dockerfile generation (optimized for each type)
- ✅ .dockerignore generation
- ✅ Platform recommendations
- ✅ CLI structure and commands

### Next
- [ ] Vercel deployment integration
- [ ] Railway deployment integration
- [ ] Render deployment integration
- [ ] OAuth flow for platform authentication
- [ ] Environment variable sync
- [ ] Deployment progress tracking

### Future
- [ ] Monorepo support (interactive selection)
- [ ] Multi-project deployment
- [ ] Rollback commands
- [ ] Deployment history
- [ ] Web dashboard (paid feature)
- [ ] Team collaboration (paid feature)

## Why Dockerfile-Based Deployment?

We chose Dockerfile-based deployment because:

1. **Universal** - Every platform supports Docker
2. **Reproducible** - Same build everywhere
3. **Optimized** - Multi-stage builds minimize image size
4. **Flexible** - Users can customize the Dockerfile
5. **No lock-in** - Switch platforms without changing code

The platform handles the Docker build, so users don't need Docker installed locally.

## Contributing

Contributions welcome! Areas we need help with:

- Platform-specific deployment logic (Vercel API, Railway API, etc.)
- More Dockerfile templates for edge cases
- Better monorepo detection
- Testing across different project types
- Documentation and examples

## License

MIT

## Links

- Website: https://myshipmate.cc
- GitHub: https://github.com/myshipmate-cli/myshipmate

---

Built with ❤️ for developers who just want to ship.
