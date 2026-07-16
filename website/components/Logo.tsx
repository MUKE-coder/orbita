export function Logo({ size = 30 }: { size?: number }) {
  return (
    <svg width={size} height={size} viewBox="0 0 48 48" fill="none" aria-hidden>
      <circle cx="24" cy="24" r="7" fill="#f45b48" />
      <ellipse cx="24" cy="24" rx="20" ry="9" stroke="#f45b48" strokeWidth="2" fill="none" transform="rotate(-24 24 24)" />
      <ellipse cx="24" cy="24" rx="20" ry="9" stroke="#f6836f" strokeWidth="2" fill="none" opacity="0.45" transform="rotate(28 24 24)" />
      <circle cx="42" cy="16" r="2.4" fill="#f6836f" />
    </svg>
  )
}
