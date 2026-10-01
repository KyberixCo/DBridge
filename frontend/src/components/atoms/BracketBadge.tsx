import React from 'react'

interface BracketBadgeProps {
  label: string
  variant?: 'acid' | 'mint' | 'purple' | 'amber' | 'cyan' | 'danger' | 'dim' | 'solid-acid' | 'solid-crimson'
  pulse?: boolean
}

export const BracketBadge: React.FC<BracketBadgeProps> = ({
  label,
  variant = 'acid',
  pulse = false,
}) => {
  if (variant === 'solid-acid') {
    return (
      <span className="inline-flex items-center gap-1.5 px-2 py-0.5 font-mono text-[10px] font-black uppercase bg-[#d9ff3f] text-[#050505]">
        {pulse && <span className="w-1.5 h-1.5 bg-[#050505] animate-ping shrink-0" />}
        <span>{label}</span>
      </span>
    )
  }

  if (variant === 'solid-crimson') {
    return (
      <span className="inline-flex items-center gap-1.5 px-2 py-0.5 font-mono text-[10px] font-black uppercase bg-[#e11d48] text-white">
        {pulse && <span className="w-1.5 h-1.5 bg-white animate-ping shrink-0" />}
        <span>{label}</span>
      </span>
    )
  }

  const colorClasses = {
    acid: "text-[#d9ff3f] border-[#383838] bg-[#0c0c0c]",
    mint: "text-[#00ea8d] border-[#383838] bg-[#0c0c0c]",
    purple: "text-[#a855f7] border-[#383838] bg-[#0c0c0c]",
    amber: "text-[#fbbf24] border-[#383838] bg-[#0c0c0c]",
    cyan: "text-[#38bdf8] border-[#383838] bg-[#0c0c0c]",
    danger: "text-[#f87171] border-[#383838] bg-[#0c0c0c]",
    dim: "text-[#a7a49c] border-[#262626] bg-[#0c0c0c]",
  }[variant] || "text-[#d9ff3f] border-[#383838] bg-[#0c0c0c]"

  return (
    <span className={`inline-flex items-center gap-1.5 px-2 py-0.5 font-mono text-[10px] font-extrabold uppercase border ${colorClasses}`}>
      {pulse && (
        <span className="w-1.5 h-1.5 rounded-full bg-current animate-pulse shrink-0" />
      )}
      <span>[ {label} ]</span>
    </span>
  )
}
