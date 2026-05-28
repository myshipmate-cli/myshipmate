export default function Home() {
  return (
    <main className="min-h-screen bg-gradient-to-b from-slate-900 via-slate-800 to-slate-900 text-white">
      {/* Navbar */}
      <nav className="border-b border-slate-700/50">
        <div className="max-w-6xl mx-auto px-4 py-4 flex justify-between items-center">
          <div className="flex items-center gap-2">
            <span className="text-2xl">🚀</span>
            <span className="text-xl font-bold">Shipmate</span>
          </div>
          <div className="flex items-center gap-6 text-sm text-slate-400">
            <a href="#features" className="hover:text-white transition-colors">Features</a>
            <a href="#platforms" className="hover:text-white transition-colors">Platforms</a>
            <a href="#languages" className="hover:text-white transition-colors">Languages</a>
            <a href="https://github.com/shipmate/cli" className="hover:text-white transition-colors">GitHub</a>
          </div>
        </div>
      </nav>

      <div className="max-w-6xl mx-auto px-4 py-20">
        {/* Hero Section */}
        <div className="text-center mb-20">
          <div className="inline-block bg-emerald-500/10 text-emerald-400 text-sm px-4 py-1.5 rounded-full mb-6 border border-emerald-500/20">
            v0.1.0 — Now with Heroku support
          </div>
          <h1 className="text-5xl md:text-7xl font-bold mb-6 bg-gradient-to-r from-white to-slate-400 bg-clip-text text-transparent">
            Deploy anywhere.<br />One command.
          </h1>
          <p className="text-xl text-slate-400 max-w-2xl mx-auto mb-12">
            Shipmate detects your project type, generates an optimized Dockerfile, and deploys to the right platform — all with a single command.
          </p>

          {/* Install Command */}
          <div className="bg-slate-950 rounded-xl p-6 inline-block mb-8 border border-slate-700/50 shadow-2xl">
            <div className="flex items-center gap-4">
              <code className="text-green-400 text-lg font-mono">
                $ curl -fsSL myshipmate.fly.dev/install | sh
              </code>
            </div>
          </div>
          <p className="text-sm text-slate-500">Or install with Go: <code className="text-slate-400">go install github.com/shipmate/cli@latest</code></p>
        </div>

        {/* How It Works */}
        <div className="mb-24">
          <h2 className="text-3xl md:text-4xl font-bold mb-4 text-center">How It Works</h2>
          <p className="text-slate-400 text-center mb-12 max-w-xl mx-auto">Three steps. Zero configuration. No Docker required.</p>
          <div className="grid md:grid-cols-3 gap-8">
            <div className="bg-slate-800/50 p-8 rounded-xl border border-slate-700/50">
              <div className="w-12 h-12 bg-blue-500/20 rounded-lg flex items-center justify-center text-2xl mb-4">🔍</div>
              <h3 className="text-xl font-bold mb-3">1. Detect</h3>
              <p className="text-slate-400">
                Scans your project and detects 30+ frameworks and languages automatically — Next.js, Go, Django, Rails, Spring Boot, and more.
              </p>
            </div>
            <div className="bg-slate-800/50 p-8 rounded-xl border border-slate-700/50">
              <div className="w-12 h-12 bg-purple-500/20 rounded-lg flex items-center justify-center text-2xl mb-4">🎯</div>
              <h3 className="text-xl font-bold mb-3">2. Recommend</h3>
              <p className="text-slate-400">
                Based on your project type, recommends the best deployment platforms. Generates an optimized Dockerfile and .dockerignore.
              </p>
            </div>
            <div className="bg-slate-800/50 p-8 rounded-xl border border-slate-700/50">
              <div className="w-12 h-12 bg-emerald-500/20 rounded-lg flex items-center justify-center text-2xl mb-4">🚀</div>
              <h3 className="text-xl font-bold mb-3">3. Deploy</h3>
              <p className="text-slate-400">
                Authenticates, syncs your env vars, uploads your code, and deploys. You get a live URL in under a minute.
              </p>
            </div>
          </div>
        </div>

        {/* Example Usage */}
        <div className="mb-24">
          <h2 className="text-3xl md:text-4xl font-bold mb-4 text-center">See It In Action</h2>
          <p className="text-slate-400 text-center mb-12">Just cd into your project and run <code className="bg-slate-800 px-2 py-0.5 rounded text-emerald-400">shipmate</code></p>
          <div className="bg-slate-950 rounded-xl p-8 max-w-3xl mx-auto border border-slate-700/50 shadow-2xl">
            <div className="flex items-center gap-2 mb-4">
              <div className="w-3 h-3 rounded-full bg-red-500/80"></div>
              <div className="w-3 h-3 rounded-full bg-yellow-500/80"></div>
              <div className="w-3 h-3 rounded-full bg-green-500/80"></div>
              <span className="text-slate-500 text-xs ml-2 font-mono">Terminal</span>
            </div>
            <pre className="text-sm overflow-x-auto font-mono leading-relaxed">
              <code className="text-slate-300">{`$ cd my-flask-app
$ shipmate

  ╔═══════════════════════════════════════╗
  ║       🚀 SHIPMATE v0.1.0              ║
  ║   The Smart Deployer for Developers   ║
  ╚═══════════════════════════════════════╝

🔍 Project detected:
   ✓ Project: my-flask-app
   ✓ Type: flask
   ✓ Build:
   ✓ Start: gunicorn app:app
   ✓ Port: 8000
   ✓ Environment file detected

   📋 Found 3 environment variables:
      • DATABASE_URL = po********db
      • SECRET_KEY   = su*********et
      • FLASK_ENV    = pr********on

📦 Generating Dockerfile...
   ✓ Dockerfile created
   ✓ .dockerignore created

ℹ️  No local Docker needed — platforms build your Dockerfile

🎯 Recommended platforms:
   ❯[1] Railway
    [2] Render
    [3] Heroku
    [4] Fly.io

   Select platform (number): 3

🔐 You need to log in to heroku first.
   ✓ Opened: https://dashboard.heroku.com/account
   ✓ Token saved for heroku

🚀 Deploying to Heroku...
   ✓ Authenticated with Heroku CLI
   ✓ App created: my-flask-app
   ✓ Set 3 config vars
   ✓ Buildpack: heroku/python
   📤 Deploying via git push...

   ✓ Deployed successfully!
   🌐 Live at: https://my-flask-app.herokuapp.com

✓ Done in 23.4s`}</code>
            </pre>
          </div>
        </div>

        {/* Supported Platforms */}
        <div id="platforms" className="mb-24">
          <h2 className="text-3xl md:text-4xl font-bold mb-4 text-center">Supported Platforms</h2>
          <p className="text-slate-400 text-center mb-12">Deploy to any of these platforms with one command</p>
          <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-4">
            {[
              { name: 'Vercel', icon: '▲', desc: 'Next.js, React, Static' },
              { name: 'Railway', icon: '🚂', desc: 'Go, Node, Python, Docker' },
              { name: 'Render', icon: '🎨', desc: 'Go, Node, Python, Ruby' },
              { name: 'Netlify', icon: '◆', desc: 'Static, JAMstack' },
              { name: 'Fly.io', icon: '✈️', desc: 'Docker, Go, Node' },
              { name: 'Heroku', icon: '💜', desc: 'All languages' },
            ].map((p) => (
              <div key={p.name} className="bg-slate-800/50 p-5 rounded-xl border border-slate-700/50 text-center hover:border-slate-600 transition-colors">
                <div className="text-3xl mb-2">{p.icon}</div>
                <div className="font-semibold text-sm">{p.name}</div>
                <div className="text-xs text-slate-500 mt-1">{p.desc}</div>
              </div>
            ))}
          </div>
        </div>

        {/* Supported Languages */}
        <div id="languages" className="mb-24">
          <h2 className="text-3xl md:text-4xl font-bold mb-4 text-center">30+ Languages & Frameworks</h2>
          <p className="text-slate-400 text-center mb-12">Shipmate detects your stack and generates optimized Dockerfiles</p>
          <div className="grid md:grid-cols-3 gap-6">
            <div className="bg-slate-800/50 p-6 rounded-xl border border-slate-700/50">
              <h3 className="font-bold text-blue-400 mb-3">JavaScript / TypeScript</h3>
              <div className="flex flex-wrap gap-2">
                {['Next.js', 'React', 'Vue', 'Angular', 'Svelte', 'SvelteKit', 'Nuxt', 'Remix', 'Astro', 'TanStack', 'Gatsby', 'SolidJS', 'Qwik', 'Node.js'].map((lang) => (
                  <span key={lang} className="bg-slate-700/50 px-2 py-1 rounded text-xs text-slate-300">{lang}</span>
                ))}
              </div>
            </div>
            <div className="bg-slate-800/50 p-6 rounded-xl border border-slate-700/50">
              <h3 className="font-bold text-emerald-400 mb-3">Backend Languages</h3>
              <div className="flex flex-wrap gap-2">
                {['Go', 'Python', 'Ruby', 'Java', 'Kotlin', 'Rust', 'C#/.NET', 'PHP', 'Elixir', 'Dart'].map((lang) => (
                  <span key={lang} className="bg-slate-700/50 px-2 py-1 rounded text-xs text-slate-300">{lang}</span>
                ))}
              </div>
            </div>
            <div className="bg-slate-800/50 p-6 rounded-xl border border-slate-700/50">
              <h3 className="font-bold text-purple-400 mb-3">Frameworks</h3>
              <div className="flex flex-wrap gap-2">
                {['Django', 'Flask', 'FastAPI', 'Rails', 'Sinatra', 'Spring Boot', 'Laravel', 'Symfony', 'Express', 'Flutter Web'].map((lang) => (
                  <span key={lang} className="bg-slate-700/50 px-2 py-1 rounded text-xs text-slate-300">{lang}</span>
                ))}
              </div>
            </div>
          </div>
        </div>

        {/* Features */}
        <div id="features" className="mb-24">
          <h2 className="text-3xl md:text-4xl font-bold mb-4 text-center">Features</h2>
          <p className="text-slate-400 text-center mb-12">Everything you need to deploy without the headache</p>
          <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-6 max-w-5xl mx-auto">
            {[
              { title: 'Smart Detection', desc: 'Detects 30+ project types from file signatures. No configuration needed.', icon: '🔍' },
              { title: 'No Docker Required', desc: 'Platforms build your Dockerfile in their cloud. No local Docker needed.', icon: '☁️' },
              { title: 'Monorepo Support', desc: 'Turborepo, Nx, pnpm workspaces, Lerna, and Yarn workspaces.', icon: '📦' },
              { title: 'Env Var Sync', desc: 'Reads .env files and pushes them to your deployment platform securely.', icon: '🔐' },
              { title: 'Optimized Dockerfiles', desc: 'Multi-stage builds, minimal images, production-ready defaults.', icon: '⚡' },
              { title: 'Platform Switching', desc: 'Try Railway, then switch to Fly.io. One command, zero lock-in.', icon: '🔄' },
              { title: 'One Command', desc: 'No config files, no YAML, no setup. Just run shipmate.', icon: '✨' },
              { title: 'Smart Recommendations', desc: 'Recommends the best platform for your specific project type.', icon: '🎯' },
              { title: 'Deploy History', desc: 'Tracks your deployments and lets you redeploy with one command.', icon: '📋' },
            ].map((f) => (
              <div key={f.title} className="bg-slate-800/50 p-6 rounded-xl border border-slate-700/50">
                <div className="text-2xl mb-3">{f.icon}</div>
                <h3 className="font-bold mb-2">{f.title}</h3>
                <p className="text-slate-400 text-sm">{f.desc}</p>
              </div>
            ))}
          </div>
        </div>

        {/* CLI Commands */}
        <div className="mb-24">
          <h2 className="text-3xl md:text-4xl font-bold mb-4 text-center">CLI Commands</h2>
          <p className="text-slate-400 text-center mb-12">Simple, intuitive commands</p>
          <div className="bg-slate-950 rounded-xl p-8 max-w-2xl mx-auto border border-slate-700/50">
            <div className="space-y-4 font-mono text-sm">
              <div className="flex gap-4">
                <code className="text-emerald-400 whitespace-nowrap">shipmate</code>
                <span className="text-slate-400">Full deploy — detect, build, and ship</span>
              </div>
              <div className="flex gap-4">
                <code className="text-emerald-400 whitespace-nowrap">shipmate init</code>
                <span className="text-slate-400">Generate Dockerfile only (no deploy)</span>
              </div>
              <div className="flex gap-4">
                <code className="text-emerald-400 whitespace-nowrap">shipmate login vercel</code>
                <span className="text-slate-400">Authenticate with a platform</span>
              </div>
              <div className="flex gap-4">
                <code className="text-emerald-400 whitespace-nowrap">shipmate status</code>
                <span className="text-slate-400">Show project info and auth status</span>
              </div>
              <div className="flex gap-4">
                <code className="text-emerald-400 whitespace-nowrap">shipmate env</code>
                <span className="text-slate-400">Show detected env vars (masked)</span>
              </div>
              <div className="flex gap-4">
                <code className="text-emerald-400 whitespace-nowrap">shipmate logout</code>
                <span className="text-slate-400">Remove stored credentials</span>
              </div>
            </div>
          </div>
        </div>

        {/* CTA */}
        <div className="text-center mb-20 bg-gradient-to-r from-blue-500/10 to-purple-500/10 rounded-2xl p-12 border border-slate-700/50">
          <h2 className="text-3xl md:text-4xl font-bold mb-4">Ready to ship?</h2>
          <p className="text-slate-400 mb-8 max-w-lg mx-auto">
            Stop wasting time on deployment configuration. Shipmate handles it all so you can focus on building.
          </p>
          <div className="bg-slate-950 rounded-xl p-4 inline-block border border-slate-700/50">
            <code className="text-green-400 font-mono">
              $ curl -fsSL myshipmate.fly.dev/install | sh
            </code>
          </div>
        </div>

        {/* Footer */}
        <footer className="text-center text-slate-500 pt-12 border-t border-slate-700/50">
          <div className="flex justify-center items-center gap-2 mb-4">
            <span className="text-lg">🚀</span>
            <span className="font-bold text-slate-300">Shipmate</span>
          </div>
          <p className="mb-2">The smart deployer for developers</p>
          <p className="text-sm">© 2026 Shipmate · myshipmate.fly.dev · Open Source</p>
        </footer>
      </div>
    </main>
  );
}
