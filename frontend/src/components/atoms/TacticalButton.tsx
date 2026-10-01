import React from 'react'

interface TacticalButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'acid' | 'crimson' | 'ghost' | 'paper' | 'cyan' | 'purple'
  size?: 'sm' | 'md' | 'lg'
  icon?: React.ReactNode
}

export const TacticalButton: React.FC<TacticalButtonProps> = ({
  children,
  variant = 'acid',
  size = 'md',
  icon,
  className = '',
  disabled,
  ...props
}) => {
  const sizeClasses = {
    sm: "px-2.5 py-1 text-[11px] min-h-[30px] gap-1.5",
    md: "px-3.5 py-2 text-xs min-h-[38px] gap-2",
    lg: "px-5 py-3 text-sm min-h-[46px] gap-2.5",
  }[size]

  const variantClass = {
    acid: "brutal-button--acid",
    crimson: "brutal-button--crimson",
    ghost: "brutal-button--ghost",
    paper: "border-[#f2efe6] bg-[#f2efe6] text-[#050505] hover:box-shadow-[4px_4px_0_#383838]",
    cyan: "border-[#38bdf8] bg-[#38bdf8] text-[#050505] hover:box-shadow-[4px_4px_0_#f2efe6]",
    purple: "brutal-button--purple",
  }[variant] || "brutal-button--acid"

  return (
    <button
      disabled={disabled}
      className={`brutal-button ${variantClass} ${sizeClasses} ${className}`}
      {...props}
    >
      {icon && <span className="flex items-center shrink-0">{icon}</span>}
      <span className="font-mono font-extrabold uppercase tracking-wider">{children}</span>
    </button>
  )
}
