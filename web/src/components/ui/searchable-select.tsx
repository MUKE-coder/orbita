import { useEffect, useLayoutEffect, useMemo, useRef, useState } from "react";
import { createPortal } from "react-dom";
import { Check, ChevronsUpDown, Search, X } from "lucide-react";

import { cn } from "@/lib/utils";

export interface SearchableOption {
  value: string;
  label: string;
  hint?: string; // smaller text to the right
  group?: string;
  disabled?: boolean;
}

interface SearchableSelectProps {
  value?: string;
  onChange: (value: string) => void;
  options: SearchableOption[];
  placeholder?: string;
  emptyText?: string;
  searchPlaceholder?: string;
  disabled?: boolean;
  className?: string;
  /** Max options rendered per open (performance cap for huge lists). */
  maxItems?: number;
  /** Custom filter function. Defaults to case-insensitive substring over label + hint. */
  filter?: (opt: SearchableOption, query: string) => boolean;
}

interface PopoverPos {
  top: number;
  left: number;
  width: number;
  /** placed above the trigger when there's no room below */
  flip: boolean;
}

/**
 * SearchableSelect — a dropdown with an inline search input. The popover is
 * rendered via React Portal at the body level so it escapes any parent with
 * `overflow: hidden`. Auto-flips above the trigger when there's no room below.
 *
 * Keyboard: Enter picks, Escape closes, Arrow keys navigate.
 */
export function SearchableSelect({
  value,
  onChange,
  options,
  placeholder = "Select...",
  emptyText = "No results",
  searchPlaceholder = "Search...",
  disabled,
  className,
  maxItems = 200,
  filter,
}: SearchableSelectProps) {
  const [open, setOpen] = useState(false);
  const [query, setQuery] = useState("");
  const [focusIdx, setFocusIdx] = useState(0);
  const [pos, setPos] = useState<PopoverPos | null>(null);

  const triggerRef = useRef<HTMLButtonElement>(null);
  const popoverRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLInputElement>(null);

  // Compute the popover position from the trigger's bounding rect.
  const reposition = () => {
    const t = triggerRef.current;
    if (!t) return;
    const rect = t.getBoundingClientRect();
    const popoverHeight = 320;
    const margin = 4;
    const spaceBelow = window.innerHeight - rect.bottom;
    const flip = spaceBelow < popoverHeight + margin && rect.top > spaceBelow;
    setPos({
      top: flip ? rect.top - margin : rect.bottom + margin,
      left: rect.left,
      width: rect.width,
      flip,
    });
  };

  // Close on outside click (including the body-rendered popover)
  useEffect(() => {
    if (!open) return;
    const handler = (e: MouseEvent) => {
      const target = e.target as Node;
      if (
        !triggerRef.current?.contains(target) &&
        !popoverRef.current?.contains(target)
      ) {
        setOpen(false);
      }
    };
    document.addEventListener("mousedown", handler);
    return () => document.removeEventListener("mousedown", handler);
  }, [open]);

  // Recompute position on open + on window resize/scroll while open
  useLayoutEffect(() => {
    if (!open) return;
    reposition();
    const onScrollResize = () => reposition();
    window.addEventListener("scroll", onScrollResize, true);
    window.addEventListener("resize", onScrollResize);
    return () => {
      window.removeEventListener("scroll", onScrollResize, true);
      window.removeEventListener("resize", onScrollResize);
    };
  }, [open]);

  // Focus the search input + reset state when opening
  useEffect(() => {
    if (open) {
      setQuery("");
      setFocusIdx(0);
      setTimeout(() => inputRef.current?.focus(), 0);
    }
  }, [open]);

  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase();
    const pred = filter
      ? (o: SearchableOption) => filter(o, q)
      : (o: SearchableOption) =>
          !q ||
          o.label.toLowerCase().includes(q) ||
          o.value.toLowerCase().includes(q) ||
          (o.hint?.toLowerCase().includes(q) ?? false);
    return options.filter(pred).slice(0, maxItems);
  }, [options, query, filter, maxItems]);

  const selected = options.find((o) => o.value === value);

  const choose = (opt: SearchableOption) => {
    if (opt.disabled) return;
    onChange(opt.value);
    setOpen(false);
  };

  const onKey = (e: React.KeyboardEvent) => {
    if (e.key === "Escape") setOpen(false);
    if (e.key === "ArrowDown") {
      e.preventDefault();
      setFocusIdx((i) => Math.min(i + 1, filtered.length - 1));
    }
    if (e.key === "ArrowUp") {
      e.preventDefault();
      setFocusIdx((i) => Math.max(i - 1, 0));
    }
    if (e.key === "Enter") {
      e.preventDefault();
      const opt = filtered[focusIdx];
      if (opt) choose(opt);
    }
  };

  return (
    <div className={cn("relative", className)}>
      <button
        ref={triggerRef}
        type="button"
        disabled={disabled}
        onClick={() => !disabled && setOpen((o) => !o)}
        className={cn(
          "flex h-10 w-full items-center justify-between gap-2 rounded-lg border border-input bg-background/50 px-3 text-left text-sm outline-none transition-colors",
          "hover:bg-accent/40",
          "focus-visible:border-ring focus-visible:ring-2 focus-visible:ring-ring/30",
          "disabled:cursor-not-allowed disabled:opacity-60",
          "dark:bg-input/40",
          open && "border-ring ring-2 ring-ring/30"
        )}
        aria-haspopup="listbox"
        aria-expanded={open}
      >
        <span
          className={cn(
            "truncate",
            !selected && "text-muted-foreground/70"
          )}
        >
          {selected ? selected.label : placeholder}
        </span>
        <ChevronsUpDown className="h-3.5 w-3.5 shrink-0 text-muted-foreground" />
      </button>

      {open && pos &&
        createPortal(
          <div
            ref={popoverRef}
            style={{
              position: "fixed",
              top: pos.flip ? undefined : pos.top,
              bottom: pos.flip ? window.innerHeight - pos.top : undefined,
              left: pos.left,
              width: pos.width,
              zIndex: 80,
            }}
            className="overflow-hidden rounded-lg border border-border bg-popover text-popover-foreground shadow-xl animate-in fade-in-0 zoom-in-95"
            onKeyDown={onKey}
          >
            <div className="relative border-b border-border">
              <Search className="pointer-events-none absolute left-3 top-1/2 h-3.5 w-3.5 -translate-y-1/2 text-muted-foreground" />
              <input
                ref={inputRef}
                value={query}
                onChange={(e) => {
                  setQuery(e.target.value);
                  setFocusIdx(0);
                }}
                placeholder={searchPlaceholder}
                className="h-10 w-full bg-transparent pl-9 pr-8 text-sm outline-none placeholder:text-muted-foreground/60"
              />
              {query && (
                <button
                  type="button"
                  onClick={() => setQuery("")}
                  className="absolute right-2 top-1/2 inline-flex h-5 w-5 -translate-y-1/2 items-center justify-center rounded-sm text-muted-foreground hover:bg-accent hover:text-foreground"
                  aria-label="Clear"
                >
                  <X className="h-3 w-3" />
                </button>
              )}
            </div>

            <div className="max-h-72 overflow-y-auto py-1">
              {filtered.length === 0 ? (
                <div className="px-3 py-6 text-center text-xs text-muted-foreground">
                  {emptyText}
                </div>
              ) : (
                filtered.map((opt, i) => {
                  const isActive = opt.value === value;
                  const isFocused = i === focusIdx;
                  return (
                    <button
                      key={opt.value}
                      type="button"
                      disabled={opt.disabled}
                      onMouseEnter={() => setFocusIdx(i)}
                      onClick={() => choose(opt)}
                      className={cn(
                        "flex w-full items-center gap-2 px-3 py-1.5 text-left text-sm transition-colors",
                        "disabled:cursor-not-allowed disabled:opacity-50",
                        isFocused && "bg-accent",
                        isActive && "text-brand"
                      )}
                    >
                      {isActive ? (
                        <Check className="h-3.5 w-3.5 shrink-0 text-brand" />
                      ) : (
                        <span className="h-3.5 w-3.5 shrink-0" />
                      )}
                      <span className="flex-1 truncate">{opt.label}</span>
                      {opt.hint && (
                        <span className="shrink-0 text-[11px] text-muted-foreground">
                          {opt.hint}
                        </span>
                      )}
                    </button>
                  );
                })
              )}
              {filtered.length === maxItems && options.length > maxItems && (
                <div className="border-t border-border px-3 py-1.5 text-[11px] text-muted-foreground">
                  Showing first {maxItems} · refine search to narrow
                </div>
              )}
            </div>
          </div>,
          document.body
        )}
    </div>
  );
}

export default SearchableSelect;
