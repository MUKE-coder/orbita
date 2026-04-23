import { Link } from "react-router-dom";
import { ReactNode } from "react";

import Logo from "@/components/layout/Logo";
import ThemeToggle from "@/components/layout/ThemeToggle";

interface AuthShellProps {
  title: string;
  description?: string;
  children: ReactNode;
  footer?: ReactNode;
}

export function AuthShell({ title, description, children, footer }: AuthShellProps) {
  return (
    <div className="relative flex min-h-screen flex-col bg-background">
      {/* Ambient glow */}
      <div className="pointer-events-none absolute inset-0 bg-radial-glow" aria-hidden />
      <div className="pointer-events-none absolute inset-0 bg-grid opacity-40" aria-hidden />

      {/* Top bar */}
      <header className="relative z-10 flex h-14 items-center justify-between px-6">
        <Logo size="md" to="/" />
        <div className="flex items-center gap-2">
          <ThemeToggle />
        </div>
      </header>

      {/* Content */}
      <main className="relative z-10 flex flex-1 items-center justify-center px-6 py-12">
        <div className="w-full max-w-[420px]">
          <div className="mb-8 text-center">
            <h1 className="font-heading text-[28px] font-semibold tracking-tight text-foreground">
              {title}
            </h1>
            {description && (
              <p className="mt-2 text-[15px] text-muted-foreground">{description}</p>
            )}
          </div>

          <div className="rounded-xl border border-border bg-card p-6 shadow-lg">
            {children}
          </div>

          {footer && (
            <div className="mt-6 text-center text-sm text-muted-foreground">
              {footer}
            </div>
          )}
        </div>
      </main>

      {/* Footer */}
      <footer className="relative z-10 border-t border-border/60 bg-background/50 py-4 backdrop-blur">
        <div className="mx-auto flex max-w-6xl flex-wrap items-center justify-between gap-3 px-6 text-xs text-muted-foreground">
          <span>&copy; {new Date().getFullYear()} Orbita. Self-hosted PaaS.</span>
          <div className="flex items-center gap-4">
            <Link to="/docs" className="transition-colors hover:text-foreground">
              Docs
            </Link>
            <a
              href="https://github.com/MUKE-coder/orbita"
              target="_blank"
              rel="noreferrer"
              className="transition-colors hover:text-foreground"
            >
              GitHub
            </a>
          </div>
        </div>
      </footer>
    </div>
  );
}

export default AuthShell;
