export function Logo({ size = 30 }: { size?: number }) {
  return (
    <svg width={size} height={size} viewBox="0 0 48 48" fill="none" aria-hidden>
      <circle cx="24" cy="24" r="7" fill="#6D5CE7" />
      <ellipse cx="24" cy="24" rx="20" ry="9" stroke="#22D3EE" strokeWidth="2" fill="none" transform="rotate(-24 24 24)" />
      <ellipse cx="24" cy="24" rx="20" ry="9" stroke="#6D5CE7" strokeWidth="2" fill="none" opacity="0.55" transform="rotate(28 24 24)" />
      <circle cx="42" cy="16" r="2.4" fill="#22D3EE" />
    </svg>
  )
}
