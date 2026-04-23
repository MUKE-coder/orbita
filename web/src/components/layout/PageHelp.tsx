import { useState, type ReactNode } from "react";
import { Link } from "react-router-dom";
import { Info, ArrowRight, HelpCircle, X } from "lucide-react";

import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { cn } from "@/lib/utils";

export interface HelpStep {
  title: string;
  body: ReactNode;
}

export interface HelpLink {
  label: string;
  to: string;
  description?: string;
}

interface PageHelpProps {
  /** Short title — the name of the current page. */
  title: string;
  /** One-line summary rendered in the trigger tooltip + modal header. */
  summary: string;
  /** Ordered list of "getting started" steps for this page. */
  steps?: HelpStep[];
  /** "Where to go next" — related pages in the app. */
  nextLinks?: HelpLink[];
  /** Optional extra free-form content rendered below the steps. */
  children?: ReactNode;
  /** Visual variant of the trigger button. Defaults to ghost icon-only. */
  variant?: "icon" | "inline";
  className?: string;
}

/**
 * PageHelp renders a small info button that opens a modal explaining the
 * current page, how to get started, and where to go next. Drop into any
 * page header.
 */
export function PageHelp({
  title,
  summary,
  steps,
  nextLinks,
  children,
  variant = "icon",
  className,
}: PageHelpProps) {
  const [open, setOpen] = useState(false);

  const trigger =
    variant === "inline" ? (
      <Button
        variant="outline"
        size="sm"
        onClick={() => setOpen(true)}
        className={cn("gap-1.5", className)}
      >
        <HelpCircle className="h-3.5 w-3.5" />
        How this page works
      </Button>
    ) : (
      <button
        type="button"
        onClick={() => setOpen(true)}
        className={cn(
          "inline-flex h-7 w-7 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-accent hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring/30",
          className
        )}
        aria-label={`About ${title}`}
        title={summary}
      >
        <Info className="h-4 w-4" />
      </button>
    );

  return (
    <>
      {trigger}
      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent className="max-w-xl">
          <DialogHeader>
            <div className="flex items-start gap-3">
              <div className="mt-0.5 flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-brand/10 text-brand">
                <Info className="h-4 w-4" />
              </div>
              <div className="min-w-0 flex-1">
                <DialogTitle className="font-heading text-lg font-semibold tracking-tight">
                  {title}
                </DialogTitle>
                <DialogDescription className="mt-1 text-sm text-muted-foreground">
                  {summary}
                </DialogDescription>
              </div>
              <button
                type="button"
                onClick={() => setOpen(false)}
                className="ml-2 inline-flex h-7 w-7 shrink-0 items-center justify-center rounded-md text-muted-foreground hover:bg-accent hover:text-foreground"
                aria-label="Close"
              >
                <X className="h-4 w-4" />
              </button>
            </div>
          </DialogHeader>

          {steps && steps.length > 0 && (
            <div className="space-y-3">
              <div className="text-[11px] font-medium uppercase tracking-wider text-muted-foreground">
                Getting started
              </div>
              <ol className="space-y-3">
                {steps.map((step, i) => (
                  <li key={step.title} className="flex gap-3">
                    <div className="flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-muted text-[11px] font-semibold text-foreground">
                      {i + 1}
                    </div>
                    <div className="min-w-0 flex-1">
                      <div className="text-sm font-medium">{step.title}</div>
                      <div className="mt-0.5 text-[13px] leading-relaxed text-muted-foreground">
                        {step.body}
                      </div>
                    </div>
                  </li>
                ))}
              </ol>
            </div>
          )}

          {children && <div className="text-sm text-muted-foreground">{children}</div>}

          {nextLinks && nextLinks.length > 0 && (
            <div className="space-y-2">
              <div className="text-[11px] font-medium uppercase tracking-wider text-muted-foreground">
                Where to go next
              </div>
              <div className="space-y-1">
                {nextLinks.map((link) => (
                  <Link
                    key={link.to}
                    to={link.to}
                    onClick={() => setOpen(false)}
                    className="group flex items-center gap-3 rounded-lg border border-border bg-card px-3 py-2 transition-colors hover:border-brand/40 hover:bg-accent/50"
                  >
                    <div className="min-w-0 flex-1">
                      <div className="text-sm font-medium">{link.label}</div>
                      {link.description && (
                        <div className="text-xs text-muted-foreground">
                          {link.description}
                        </div>
                      )}
                    </div>
                    <ArrowRight className="h-3.5 w-3.5 text-muted-foreground transition-colors group-hover:text-brand" />
                  </Link>
                ))}
              </div>
            </div>
          )}

          <div className="flex justify-end border-t border-border pt-3">
            <Button variant="ghost" size="sm" onClick={() => setOpen(false)}>
              Got it
            </Button>
          </div>
        </DialogContent>
      </Dialog>
    </>
  );
}

export default PageHelp;
