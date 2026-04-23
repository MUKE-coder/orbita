import { Link } from "react-router-dom";
import { cn } from "@/lib/utils";

interface LogoProps {
  className?: string;
  iconOnly?: boolean;
  to?: string;
  size?: "sm" | "md" | "lg";
}

export function Logo({ className, iconOnly = false, to = "/", size = "md" }: LogoProps) {
  const sizes = {
    sm: { icon: "h-6 w-6", text: "text-sm" },
    md: { icon: "h-7 w-7", text: "text-base" },
    lg: { icon: "h-10 w-10", text: "text-xl" },
  };
  const s = sizes[size];

  const inner = (
    <div className={cn("flex items-center gap-2.5", className)}>
      <div className={cn("relative flex-shrink-0", s.icon)}>
        <svg viewBox="0 0 32 32" className={cn(s.icon, "text-brand")} fill="none">
          <defs>
            <linearGradient id="orbita-grad" x1="0" y1="0" x2="32" y2="32">
              <stop offset="0%" stopColor="currentColor" stopOpacity="1" />
              <stop offset="100%" stopColor="currentColor" stopOpacity="0.6" />
            </linearGradient>
          </defs>
          <circle cx="16" cy="16" r="13" stroke="url(#orbita-grad)" strokeWidth="1.5" fill="none" />
          <ellipse
            cx="16"
            cy="16"
            rx="13"
            ry="5"
            stroke="url(#orbita-grad)"
            strokeWidth="1.5"
            fill="none"
            transform="rotate(-25 16 16)"
          />
          <circle cx="16" cy="16" r="3.5" fill="currentColor" />
        </svg>
      </div>
      {!iconOnly && (
        <span className={cn("font-semibold tracking-tight text-foreground", s.text)}>
          Orbita
        </span>
      )}
    </div>
  );

  return to ? <Link to={to}>{inner}</Link> : inner;
}

export default Logo;
