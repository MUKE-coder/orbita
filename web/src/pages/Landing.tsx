import { Link } from "react-router-dom";
import {
  Rocket,
  Shield,
  Server,
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
  ArrowRight,
  GitFork,
} from "lucide-react";

import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from "@/components/ui/accordion";

const features = [
  {
    icon: Users,
    title: "True Multi-Tenancy",
    description:
      "Fully isolated organizations with separate networks, secrets, volumes, and resource quotas. Clients never see each other.",
  },
  {
    icon: Container,
    title: "Docker-Native Deployment",
    description:
      "Deploy from Docker images, Git repos, or Docker Compose files. Zero-downtime rolling deploys with rollback support.",
  },
  {
    icon: Shield,
    title: "Resource Isolation",
    description:
      "Per-org cgroup slices enforce CPU and memory limits. No noisy neighbor problems — every tenant gets their fair share.",
  },
  {
    icon: Database,
    title: "Managed Databases",
    description:
      "One-click PostgreSQL, MySQL, MariaDB, MongoDB, and Redis. Auto-generated passwords, encrypted connection strings, scheduled backups.",
  },
  {
    icon: Globe,
    title: "Automatic SSL",
    description:
      "Custom domains with automatic Let's Encrypt TLS via Traefik. HTTP to HTTPS redirect, wildcard subdomain support.",
  },
  {
    icon: GitBranch,
    title: "Git Auto-Deploy",
    description:
      "Connect GitHub, GitLab, or Gitea. Auto-deploy on push via webhooks. Build from Dockerfile or Nixpacks.",
  },
  {
    icon: Clock,
    title: "Cron Jobs",
    description:
      "Scheduled Docker containers that run and exit. Concurrency policies, timeout enforcement, run history with logs.",
  },
  {
    icon: Terminal,
    title: "In-Browser Terminal",
    description:
      "SSH into running containers directly from the browser using xterm.js. Audit-logged for compliance.",
  },
  {
    icon: BarChart3,
    title: "Real-Time Monitoring",
    description:
      "CPU, memory, network, and disk metrics per container. Dashboard overview with resource usage bars.",
  },
  {
    icon: Layers,
    title: "Service Marketplace",
    description:
      "One-click deploy WordPress, Plausible, n8n, Grafana, MinIO, Vaultwarden, and more from pre-built templates.",
  },
  {
    icon: Lock,
    title: "Encrypted Secrets",
    description:
      "AES-256-GCM encryption with per-org derived keys via HKDF. Secrets masked in UI, redacted in logs.",
  },
  {
    icon: Server,
    title: "Multi-Node Swarm",
    description:
      "Add worker nodes via SSH. Docker Swarm orchestration with node labels, drain, and placement constraints.",
  },
];

const comparisonData = [
  { feature: "Multi-tenancy", orbita: true, dokploy: false, coolify: false },
  { feature: "Resource quotas (cgroups)", orbita: true, dokploy: false, coolify: false },
  { feature: "Working invite system", orbita: true, dokploy: false, coolify: "Partial" },
  { feature: "RBAC (4 roles)", orbita: true, dokploy: false, coolify: false },
  { feature: "Cron job manager", orbita: true, dokploy: false, coolify: false },
  { feature: "Single binary deploy", orbita: true, dokploy: false, coolify: false },
  { feature: "Idle memory usage", orbita: "<50MB", dokploy: "~200MB", coolify: "~500MB" },
  { feature: "Written in", orbita: "Go", dokploy: "Node.js", coolify: "PHP/Laravel" },
];

const faqs = [
  {
    q: "What are the minimum server requirements?",
    a: "Orbita runs on any Linux VPS with Docker installed. Minimum: 1 vCPU, 1GB RAM, 10GB disk. Recommended: 2 vCPU, 4GB RAM for hosting multiple clients.",
  },
  {
    q: "Can I migrate from Dokploy or Coolify?",
    a: "Yes. Since Orbita uses standard Docker containers, you can export your Docker images and import them into Orbita. Database backups can be restored using standard tools.",
  },
  {
    q: "Is it really free and open source?",
    a: "Yes, Orbita is 100% open source under the MIT license. No per-seat fees, no usage limits, no vendor lock-in. You own everything.",
  },
  {
    q: "How does multi-tenancy isolation work?",
    a: "Each organization gets: a dedicated Docker overlay network, org-scoped database queries, per-org AES-256 encryption keys, cgroup resource limits, and volume namespace prefixing.",
  },
  {
    q: "Does it support ARM64 servers?",
    a: "Yes, Orbita compiles for both AMD64 and ARM64. The Docker images are multi-arch.",
  },
];

function Landing() {
  return (
    <div className="min-h-screen bg-background">
      {/* Nav */}
      <nav className="border-b">
        <div className="mx-auto flex max-w-6xl items-center justify-between px-6 py-4">
          <div className="flex items-center gap-2 text-xl font-bold">
            <Rocket className="h-6 w-6" />
            Orbita
          </div>
          <div className="flex items-center gap-4">
            <Link to="/docs" className="text-sm text-muted-foreground hover:text-foreground">
              Docs
            </Link>
            <a
              href="https://github.com/MUKE-coder/orbita"
              target="_blank"
              rel="noopener noreferrer"
              className="text-sm text-muted-foreground hover:text-foreground flex items-center gap-1"
            >
              <GitFork className="h-4 w-4" /> GitHub
            </a>
            <Link to="/login">
              <Button variant="outline" size="sm">Sign In</Button>
            </Link>
            <Link to="/register">
              <Button size="sm">Get Started</Button>
            </Link>
          </div>
        </div>
      </nav>

      {/* Hero */}
      <section className="mx-auto max-w-6xl px-6 py-24 text-center">
        <Badge variant="secondary" className="mb-4">Open Source PaaS</Badge>
        <h1 className="text-5xl font-bold tracking-tight sm:text-6xl">
          One VPS. Many Clients.
          <br />
          <span className="text-primary">Full Isolation.</span>
        </h1>
        <p className="mx-auto mt-6 max-w-2xl text-lg text-muted-foreground">
          Orbita is a self-hosted Platform-as-a-Service built in Go. Turn a single server into
          a fully isolated hosting environment for multiple clients — each with their own dashboard,
          resources, domains, and secrets.
        </p>
        <div className="mt-8 flex justify-center gap-4">
          <Link to="/register">
            <Button size="lg">
              Deploy Now <ArrowRight className="ml-2 h-4 w-4" />
            </Button>
          </Link>
          <Link to="/docs">
            <Button size="lg" variant="outline">
              Read the Docs
            </Button>
          </Link>
        </div>
        <p className="mt-4 text-sm text-muted-foreground">
          Ships as a single ~30MB binary. Under 50MB RAM at idle.
        </p>
      </section>

      {/* Features */}
      <section className="border-t bg-muted/30 py-20">
        <div className="mx-auto max-w-6xl px-6">
          <h2 className="text-center text-3xl font-bold mb-4">Everything You Need to Host for Clients</h2>
          <p className="text-center text-muted-foreground mb-12 max-w-2xl mx-auto">
            Built from the ground up for freelancers, agencies, and small hosting providers who manage infrastructure for multiple clients on a single server.
          </p>
          <div className="grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
            {features.map((f) => (
              <Card key={f.title}>
                <CardHeader className="pb-2">
                  <CardTitle className="flex items-center gap-2 text-base">
                    <f.icon className="h-5 w-5 text-primary" />
                    {f.title}
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <CardDescription>{f.description}</CardDescription>
                </CardContent>
              </Card>
            ))}
          </div>
        </div>
      </section>

      {/* Comparison */}
      <section className="py-20">
        <div className="mx-auto max-w-4xl px-6">
          <h2 className="text-center text-3xl font-bold mb-4">How Orbita Compares</h2>
          <p className="text-center text-muted-foreground mb-8">
            Built to solve the problems that existing self-hosted PaaS tools haven't.
          </p>
          <div className="overflow-x-auto rounded-lg border">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b bg-muted/50">
                  <th className="px-4 py-3 text-left font-medium">Feature</th>
                  <th className="px-4 py-3 text-center font-medium text-primary">Orbita</th>
                  <th className="px-4 py-3 text-center font-medium">Dokploy</th>
                  <th className="px-4 py-3 text-center font-medium">Coolify</th>
                </tr>
              </thead>
              <tbody>
                {comparisonData.map((row) => (
                  <tr key={row.feature} className="border-b">
                    <td className="px-4 py-3">{row.feature}</td>
                    <td className="px-4 py-3 text-center">
                      {row.orbita === true ? (
                        <Check className="mx-auto h-4 w-4 text-green-500" />
                      ) : (
                        <span className="text-green-500 font-medium">{String(row.orbita)}</span>
                      )}
                    </td>
                    <td className="px-4 py-3 text-center text-muted-foreground">
                      {row.dokploy === true ? <Check className="mx-auto h-4 w-4" /> : row.dokploy === false ? "—" : String(row.dokploy)}
                    </td>
                    <td className="px-4 py-3 text-center text-muted-foreground">
                      {row.coolify === true ? <Check className="mx-auto h-4 w-4" /> : row.coolify === false ? "—" : String(row.coolify)}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      </section>

      {/* Quick Install */}
      <section className="border-t bg-muted/30 py-20">
        <div className="mx-auto max-w-3xl px-6 text-center">
          <h2 className="text-3xl font-bold mb-4">
            <Zap className="inline h-8 w-8 text-primary mr-2" />
            Install in 60 Seconds
          </h2>
          <div className="mt-8 rounded-lg bg-gray-950 p-6 text-left">
            <pre className="text-sm text-green-400 font-mono overflow-x-auto">
{`# Download and run
curl -fsSL https://raw.githubusercontent.com/MUKE-coder/orbita/main/install.sh | bash

# Or with Docker
docker run -d --name orbita \\
  -p 8080:8080 \\
  -v orbita_data:/app/data \\
  -e JWT_SECRET=your-secret-key \\
  -e JWT_REFRESH_SECRET=your-refresh-secret \\
  -e DATABASE_URL=postgres://... \\
  -e REDIS_URL=redis://... \\
  ghcr.io/muke-coder/orbita:latest`}
            </pre>
          </div>
          <p className="mt-4 text-sm text-muted-foreground">
            See the <Link to="/docs" className="text-primary hover:underline">full installation guide</Link> for detailed instructions.
          </p>
        </div>
      </section>

      {/* FAQ */}
      <section className="py-20">
        <div className="mx-auto max-w-3xl px-6">
          <h2 className="text-center text-3xl font-bold mb-8">FAQ</h2>
          <Accordion>
            {faqs.map((faq, i) => (
              <AccordionItem key={i} value={`faq-${i}`}>
                <AccordionTrigger>{faq.q}</AccordionTrigger>
                <AccordionContent>{faq.a}</AccordionContent>
              </AccordionItem>
            ))}
          </Accordion>
        </div>
      </section>

      {/* CTA */}
      <section className="border-t bg-primary/5 py-20">
        <div className="mx-auto max-w-3xl px-6 text-center">
          <h2 className="text-3xl font-bold mb-4">Ready to Host Smarter?</h2>
          <p className="text-muted-foreground mb-8">
            Stop paying per-seat fees. Stop dealing with PHP memory leaks. Deploy Orbita on your own server and start hosting clients in minutes.
          </p>
          <div className="flex justify-center gap-4">
            <Link to="/register">
              <Button size="lg">Get Started Free</Button>
            </Link>
            <a href="https://github.com/MUKE-coder/orbita" target="_blank" rel="noopener noreferrer">
              <Button size="lg" variant="outline">
                <GitFork className="mr-2 h-4 w-4" /> Star on GitHub
              </Button>
            </a>
          </div>
        </div>
      </section>

      {/* Footer */}
      <footer className="border-t py-8">
        <div className="mx-auto max-w-6xl px-6 flex items-center justify-between text-sm text-muted-foreground">
          <div className="flex items-center gap-2">
            <Rocket className="h-4 w-4" /> Orbita &mdash; Open Source PaaS
          </div>
          <div className="flex gap-4">
            <a href="https://github.com/MUKE-coder/orbita" target="_blank" rel="noopener noreferrer" className="hover:text-foreground">GitHub</a>
            <Link to="/docs" className="hover:text-foreground">Docs</Link>
            <Link to="/login" className="hover:text-foreground">Sign In</Link>
          </div>
        </div>
      </footer>
    </div>
  );
}

export default Landing;
