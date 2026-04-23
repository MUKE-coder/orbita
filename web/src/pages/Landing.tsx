import { Link } from "react-router-dom";
import {
  ArrowRight,
  Rocket,
  Shield,
  Globe,
  Database,
  Clock,
  GitBranch,
  Terminal,
  BarChart3,
  Users,
  Lock,
  Layers,
  Container,
  Zap,
  Check,
  GitFork,
  Star,
  Server,
} from "lucide-react";

import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import Logo from "@/components/layout/Logo";
import ThemeToggle from "@/components/layout/ThemeToggle";

const features = [
  {
    icon: Users,
    title: "True multi-tenancy",
    description:
      "Fully isolated organizations with separate networks, secrets, volumes, and resource quotas. Clients never see each other.",
  },
  {
    icon: Container,
    title: "Docker-native deploys",
    description:
      "Deploy from Docker images, Git repos, or Compose files. Zero-downtime rolling deploys with rollback support.",
  },
  {
    icon: Shield,
    title: "Resource isolation",
    description:
      "Per-org cgroup v2 slices enforce CPU and memory limits. No noisy neighbor problems — every tenant gets their fair share.",
  },
  {
    icon: Database,
    title: "Managed databases",
    description:
      "One-click PostgreSQL, MySQL, MariaDB, MongoDB, Redis. Auto-generated passwords, encrypted connection strings, scheduled backups.",
  },
  {
    icon: Globe,
    title: "Automatic SSL",
    description:
      "Custom domains with automatic Let's Encrypt TLS via Traefik. HTTP→HTTPS redirect, wildcard subdomain support.",
  },
  {
    icon: GitBranch,
    title: "Git auto-deploy",
    description:
      "Connect GitHub, GitLab, or Gitea. Auto-deploy on push via webhooks. Build from Dockerfile or Nixpacks.",
  },
  {
    icon: Clock,
    title: "Cron jobs",
    description:
      "Scheduled Docker containers that run and exit. Concurrency policies, timeout enforcement, run history with logs.",
  },
  {
    icon: BarChart3,
    title: "Real-time observability",
    description:
      "Live log streaming, CPU/memory/network metrics, in-browser terminal via xterm.js. Deploy history with one-click rollback.",
  },
  {
    icon: Terminal,
    title: "In-browser terminal",
    description:
      "Full xterm.js shell into any running container. No SSH setup, no port forwarding — debug production from your browser.",
  },
];

const compareRows = [
  { feature: "Multi-tenancy", orbita: true, dokploy: false, coolify: false },
  { feature: "Resource quotas (cgroups)", orbita: true, dokploy: false, coolify: false },
  { feature: "Working invite system", orbita: true, dokploy: false, coolify: "partial" },
  { feature: "RBAC (4 roles)", orbita: true, dokploy: false, coolify: false },
  { feature: "Cron job manager", orbita: true, dokploy: false, coolify: false },
  { feature: "Single binary deploy", orbita: true, dokploy: false, coolify: false },
  { feature: "Idle memory usage", orbita: "<50 MB", dokploy: "~200 MB", coolify: "~500 MB" },
  { feature: "Written in", orbita: "Go", dokploy: "Node.js", coolify: "PHP" },
];

function Landing() {
  return (
    <div className="min-h-screen bg-background text-foreground">
      {/* Top nav */}
      <header className="sticky top-0 z-50 border-b border-border/60 bg-background/80 backdrop-blur-md">
        <nav className="mx-auto flex h-14 max-w-7xl items-center justify-between px-6">
          <Logo size="md" />
          <div className="flex items-center gap-1">
            <Link to="/docs">
              <Button variant="ghost" size="sm">Docs</Button>
            </Link>
            <a
              href="https://github.com/MUKE-coder/orbita"
              target="_blank"
              rel="noreferrer"
            >
              <Button variant="ghost" size="sm">
                <GitFork className="h-3.5 w-3.5" />
                GitHub
              </Button>
            </a>
            <div className="mx-2 h-5 w-px bg-border" />
            <ThemeToggle />
            <Link to="/login">
              <Button variant="ghost" size="sm">Sign in</Button>
            </Link>
            <Link to="/register">
              <Button variant="brand" size="sm">
                Get started
                <ArrowRight className="h-3.5 w-3.5" />
              </Button>
            </Link>
          </div>
        </nav>
      </header>

      {/* Hero */}
      <section className="relative overflow-hidden border-b border-border/60">
        <div className="pointer-events-none absolute inset-0 bg-grid" aria-hidden />
        <div className="pointer-events-none absolute inset-0 bg-radial-glow" aria-hidden />

        <div className="relative mx-auto max-w-5xl px-6 pb-24 pt-20 text-center sm:pt-28">
          <div className="mb-6 inline-flex items-center gap-2 rounded-full border border-border bg-card/50 px-3 py-1 text-xs font-medium backdrop-blur">
            <span className="relative flex h-2 w-2">
              <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-brand opacity-60" />
              <span className="relative inline-flex h-2 w-2 rounded-full bg-brand" />
            </span>
            <span className="text-muted-foreground">v0.1.0 — Open Source</span>
            <div className="h-3 w-px bg-border" />
            <a
              href="https://github.com/MUKE-coder/orbita"
              target="_blank"
              rel="noreferrer"
              className="flex items-center gap-1 transition-colors hover:text-foreground"
            >
              <Star className="h-3 w-3" />
              Star on GitHub
            </a>
          </div>

          <h1 className="font-heading text-5xl font-semibold leading-[1.05] tracking-tight sm:text-6xl lg:text-7xl">
            Self-hosted PaaS
            <br />
            <span className="bg-gradient-to-r from-brand via-brand to-foreground bg-clip-text text-transparent">
              with true multi-tenancy.
            </span>
          </h1>

          <p className="mx-auto mt-6 max-w-2xl text-base leading-relaxed text-muted-foreground sm:text-lg">
            One VPS. Many clients. Full isolation. Deploy apps, databases, and cron
            jobs with per-org Docker networks, encryption keys, and resource quotas —
            from a 30 MB Go binary.
          </p>

          <div className="mt-10 flex flex-col items-center justify-center gap-3 sm:flex-row">
            <Link to="/register">
              <Button variant="brand" size="xl" className="min-w-[180px]">
                Start deploying
                <ArrowRight className="h-4 w-4" />
              </Button>
            </Link>
            <a
              href="https://github.com/MUKE-coder/orbita#-installation"
              target="_blank"
              rel="noreferrer"
            >
              <Button variant="outline" size="xl" className="min-w-[180px]">
                <Terminal className="h-4 w-4" />
                Install on your VPS
              </Button>
            </a>
          </div>

          <div className="mt-12 mx-auto max-w-2xl rounded-xl border border-border bg-card/80 p-1 font-mono text-left text-sm shadow-lg backdrop-blur">
            <div className="flex items-center gap-1.5 border-b border-border px-4 py-2">
              <span className="h-2.5 w-2.5 rounded-full bg-destructive/60" />
              <span className="h-2.5 w-2.5 rounded-full bg-warning/60" />
              <span className="h-2.5 w-2.5 rounded-full bg-success/60" />
              <span className="ml-auto text-[11px] text-muted-foreground">
                one-line install
              </span>
            </div>
            <pre className="overflow-x-auto px-4 py-3 text-[13px] leading-relaxed">
              <span className="text-muted-foreground">$ </span>
              <span className="text-foreground">curl -sSL </span>
              <span className="text-brand">
                raw.githubusercontent.com/MUKE-coder/orbita/main/install.sh
              </span>
              <span className="text-foreground"> | sudo bash</span>
            </pre>
          </div>

          {/* Stats strip */}
          <div className="mt-12 grid grid-cols-3 gap-6 border-t border-border pt-8 sm:gap-12">
            <Stat value="~30 MB" label="Single binary" />
            <Stat value="<50 MB" label="Idle RAM" />
            <Stat value="28" label="DB tables" />
          </div>
        </div>
      </section>

      {/* Feature grid */}
      <section className="border-b border-border/60 py-20 sm:py-28">
        <div className="mx-auto max-w-7xl px-6">
          <div className="mx-auto max-w-2xl text-center">
            <Badge variant="brand" className="mb-4">Core platform</Badge>
            <h2 className="font-heading text-4xl font-semibold tracking-tight sm:text-5xl">
              Everything you need.
              <br />
              Nothing you don't.
            </h2>
            <p className="mt-4 text-lg text-muted-foreground">
              A platform that runs itself. Deploy, monitor, scale — all from a
              single 30 MB binary with an embedded React dashboard.
            </p>
          </div>

          <div className="mt-16 grid gap-px overflow-hidden rounded-2xl border border-border bg-border sm:grid-cols-2 lg:grid-cols-3">
            {features.map((f) => (
              <div
                key={f.title}
                className="group flex flex-col bg-card p-6 transition-colors hover:bg-accent/30 sm:p-8"
              >
                <div className="mb-5 flex h-10 w-10 items-center justify-center rounded-lg bg-brand/10 text-brand ring-1 ring-brand/20 transition-colors group-hover:bg-brand/15">
                  <f.icon className="h-5 w-5" />
                </div>
                <h3 className="font-heading text-base font-semibold tracking-tight">
                  {f.title}
                </h3>
                <p className="mt-2 text-sm leading-relaxed text-muted-foreground">
                  {f.description}
                </p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Comparison */}
      <section className="border-b border-border/60 py-20 sm:py-28">
        <div className="mx-auto max-w-5xl px-6">
          <div className="mx-auto max-w-2xl text-center">
            <Badge variant="outline" className="mb-4">Comparison</Badge>
            <h2 className="font-heading text-4xl font-semibold tracking-tight sm:text-5xl">
              Built for agencies.
            </h2>
            <p className="mt-4 text-lg text-muted-foreground">
              Dokploy and Coolify are great for single-tenant use. Orbita is built
              from the ground up for managing many clients on one server.
            </p>
          </div>

          <div className="mx-auto mt-16 max-w-3xl overflow-hidden rounded-xl border border-border bg-card shadow-sm">
            <div className="grid grid-cols-4 border-b border-border bg-muted/40 px-6 py-4 text-sm font-medium">
              <div />
              <div className="text-center">
                <div className="flex items-center justify-center gap-1.5">
                  <Logo iconOnly size="sm" to="" />
                  <span>Orbita</span>
                </div>
              </div>
              <div className="text-center text-muted-foreground">Dokploy</div>
              <div className="text-center text-muted-foreground">Coolify</div>
            </div>
            {compareRows.map((r, i) => (
              <div
                key={r.feature}
                className={`grid grid-cols-4 items-center px-6 py-4 text-sm ${
                  i < compareRows.length - 1 ? "border-b border-border" : ""
                }`}
              >
                <div className="font-medium">{r.feature}</div>
                <CompareCell value={r.orbita} highlight />
                <CompareCell value={r.dokploy} />
                <CompareCell value={r.coolify} />
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Security callout */}
      <section className="border-b border-border/60 py-20 sm:py-28">
        <div className="mx-auto max-w-5xl px-6">
          <div className="grid gap-10 lg:grid-cols-2 lg:items-center">
            <div>
              <Badge variant="outline" className="mb-4">
                <Lock className="h-3 w-3" />
                Security-first
              </Badge>
              <h2 className="font-heading text-4xl font-semibold tracking-tight sm:text-5xl">
                Your secrets stay yours.
              </h2>
              <p className="mt-5 text-lg leading-relaxed text-muted-foreground">
                Orbita encrypts every secret with per-organization keys derived via
                HKDF-SHA256 from your master key. The master key never leaves
                your server. Not even we can read your data — because there's no we.
              </p>
              <ul className="mt-8 space-y-3">
                {[
                  "AES-256-GCM encryption with per-org derived keys",
                  "JWT (15-min access + 30-day refresh, httpOnly)",
                  "bcrypt password hashing (cost 12)",
                  "HMAC-SHA256 webhook signature verification",
                  "Rate-limited auth endpoints (Redis sliding window)",
                  "Full audit log of every sensitive action",
                ].map((item) => (
                  <li key={item} className="flex items-start gap-3 text-sm">
                    <div className="mt-0.5 flex h-5 w-5 flex-shrink-0 items-center justify-center rounded-full bg-success/10 ring-1 ring-success/20">
                      <Check className="h-3 w-3 text-success" />
                    </div>
                    <span className="text-foreground/90">{item}</span>
                  </li>
                ))}
              </ul>
            </div>

            <div className="relative">
              <div className="absolute -inset-4 rounded-3xl bg-gradient-to-br from-brand/20 via-brand/10 to-transparent blur-2xl" />
              <div className="relative overflow-hidden rounded-2xl border border-border bg-card p-8 shadow-xl">
                <div className="mb-6 flex items-center gap-3">
                  <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-brand/10 ring-1 ring-brand/20">
                    <Layers className="h-5 w-5 text-brand" />
                  </div>
                  <div>
                    <div className="font-heading text-base font-semibold">
                      Per-tenant isolation
                    </div>
                    <div className="text-xs text-muted-foreground">
                      Enforced at every layer
                    </div>
                  </div>
                </div>

                <div className="space-y-3">
                  {[
                    { label: "Docker network", value: "org-acme_net" },
                    { label: "Encryption key", value: "HKDF(master, org.id)" },
                    { label: "cgroup slice", value: "orbita-org-acme.slice" },
                    { label: "Volume prefix", value: "acme-*" },
                    { label: "Traefik router", value: "acme-*-router" },
                  ].map((row) => (
                    <div
                      key={row.label}
                      className="flex items-center justify-between rounded-lg border border-border bg-background/50 px-3 py-2 font-mono text-xs"
                    >
                      <span className="text-muted-foreground">{row.label}</span>
                      <span className="text-foreground">{row.value}</span>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* CTA */}
      <section className="relative overflow-hidden border-b border-border/60 py-24">
        <div className="pointer-events-none absolute inset-0 bg-radial-glow" aria-hidden />
        <div className="relative mx-auto max-w-4xl px-6 text-center">
          <div className="mb-6 flex justify-center">
            <div className="flex h-14 w-14 items-center justify-center rounded-2xl border border-border bg-card shadow-lg">
              <Rocket className="h-6 w-6 text-brand" />
            </div>
          </div>
          <h2 className="font-heading text-4xl font-semibold tracking-tight sm:text-5xl">
            Deploy your first app
            <br />
            in under 60 seconds.
          </h2>
          <p className="mt-4 text-lg text-muted-foreground">
            Ubuntu 22.04+, 1 GB RAM, and your favorite domain. That's it.
          </p>
          <div className="mt-10 flex flex-col items-center justify-center gap-3 sm:flex-row">
            <Link to="/register">
              <Button variant="brand" size="xl" className="min-w-[200px]">
                <Zap className="h-4 w-4" />
                Get started free
              </Button>
            </Link>
            <a
              href="https://github.com/MUKE-coder/orbita"
              target="_blank"
              rel="noreferrer"
            >
              <Button variant="outline" size="xl" className="min-w-[200px]">
                <Server className="h-4 w-4" />
                Read the docs
              </Button>
            </a>
          </div>
        </div>
      </section>

      {/* Footer */}
      <footer className="py-10">
        <div className="mx-auto flex max-w-7xl flex-wrap items-center justify-between gap-4 px-6">
          <div className="flex items-center gap-3 text-sm text-muted-foreground">
            <Logo size="sm" />
            <span className="text-border">·</span>
            <span>&copy; {new Date().getFullYear()}</span>
            <span className="text-border">·</span>
            <span>MIT License</span>
          </div>
          <div className="flex items-center gap-5 text-sm text-muted-foreground">
            <Link to="/docs" className="transition-colors hover:text-foreground">Docs</Link>
            <a
              href="https://github.com/MUKE-coder/orbita"
              target="_blank"
              rel="noreferrer"
              className="transition-colors hover:text-foreground"
            >
              GitHub
            </a>
            <a
              href="https://github.com/MUKE-coder/orbita/issues"
              target="_blank"
              rel="noreferrer"
              className="transition-colors hover:text-foreground"
            >
              Issues
            </a>
          </div>
        </div>
      </footer>
    </div>
  );
}

function Stat({ value, label }: { value: string; label: string }) {
  return (
    <div className="text-center">
      <div className="font-heading text-2xl font-semibold tracking-tight sm:text-3xl">
        {value}
      </div>
      <div className="mt-1 text-xs text-muted-foreground">{label}</div>
    </div>
  );
}

function CompareCell({
  value,
  highlight,
}: {
  value: boolean | string;
  highlight?: boolean;
}) {
  if (typeof value === "boolean") {
    return (
      <div className="flex justify-center">
        {value ? (
          <div
            className={`flex h-6 w-6 items-center justify-center rounded-full ${
              highlight ? "bg-success/15 ring-1 ring-success/30" : "bg-success/10"
            }`}
          >
            <Check className="h-3.5 w-3.5 text-success" />
          </div>
        ) : (
          <div className="flex h-6 w-6 items-center justify-center rounded-full bg-muted/50 text-muted-foreground">
            —
          </div>
        )}
      </div>
    );
  }
  if (value === "partial") {
    return (
      <div className="text-center text-xs text-muted-foreground">Partial</div>
    );
  }
  return (
    <div
      className={`text-center text-sm font-medium ${
        highlight ? "text-brand" : "text-muted-foreground"
      }`}
    >
      {value}
    </div>
  );
}

export default Landing;
