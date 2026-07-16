// Minimal classnames joiner (dependency-free). Filters falsy values and
// joins the rest with a space — enough for the `cn(...)` call sites we use.
export function cn(...inputs: Array<string | false | null | undefined>): string {
  return inputs.filter(Boolean).join(' ')
}
