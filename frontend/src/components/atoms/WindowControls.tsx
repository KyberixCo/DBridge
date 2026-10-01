import React from 'react'
import {
  WindowMinimise,
  WindowToggleMaximise,
  Quit,
} from '../../../wailsjs/runtime/runtime'

export type PlatformType = 'darwin' | 'windows' | 'other'

interface WindowControlsProps {
  platform?: PlatformType
}

export const WindowControls: React.FC<WindowControlsProps> = ({ platform = 'darwin' }) => {
  const isMac = platform === 'darwin'

  if (isMac) {
    return (
      <div
        style={{ '--wails-draggable': 'no-drag' } as React.CSSProperties}
        className="flex items-center h-full shrink-0 border-r border-[#2e2e2e] font-mono select-none"
      >
        {/* Terminate / Close (macOS Leftmost - Red) */}
        <button
          type="button"
          onClick={() => Quit()}
          title="CERRAR // [SYS_KILL]"
          aria-label="Cerrar ventana"
          className="h-full px-3 flex items-center justify-center gap-0.5 text-[#a7a49c] hover:text-[#050505] hover:bg-[#f87171] border-r border-[#262626] transition-colors cursor-pointer group"
        >
          <span className="text-[10px] font-bold text-[#f87171]/70 group-hover:text-[#050505] opacity-60 group-hover:opacity-100">[</span>
          <span className="font-black text-xs leading-none text-[#f87171] group-hover:text-[#050505]">✕</span>
          <span className="text-[10px] font-bold text-[#f87171]/70 group-hover:text-[#050505] opacity-60 group-hover:opacity-100">]</span>
        </button>

        {/* Minimise (macOS Center - Amber/Yellow) */}
        <button
          type="button"
          onClick={() => WindowMinimise()}
          title="MINIMIZAR // [SYS_MIN]"
          aria-label="Minimizar ventana"
          className="h-full px-3 flex items-center justify-center gap-0.5 text-[#a7a49c] hover:text-[#050505] hover:bg-[#fbbf24] border-r border-[#262626] transition-colors cursor-pointer group"
        >
          <span className="text-[10px] font-bold text-[#fbbf24]/70 group-hover:text-[#050505] opacity-60 group-hover:opacity-100">[</span>
          <span className="font-black text-xs leading-none -translate-y-0.5 text-[#fbbf24] group-hover:text-[#050505]">_</span>
          <span className="text-[10px] font-bold text-[#fbbf24]/70 group-hover:text-[#050505] opacity-60 group-hover:opacity-100">]</span>
        </button>

        {/* Toggle Maximise (macOS Rightmost - Mint/Green) */}
        <button
          type="button"
          onClick={() => WindowToggleMaximise()}
          title="MAXIMIZAR / RESTAURAR // [SYS_EXPAND]"
          aria-label="Maximizar o restaurar ventana"
          className="h-full px-3 flex items-center justify-center gap-0.5 text-[#a7a49c] hover:text-[#050505] hover:bg-[#00ea8d] transition-colors cursor-pointer group"
        >
          <span className="text-[10px] font-bold text-[#00ea8d]/70 group-hover:text-[#050505] opacity-60 group-hover:opacity-100">[</span>
          <span className="font-black text-[11px] leading-none text-[#00ea8d] group-hover:text-[#050505]">⛶</span>
          <span className="text-[10px] font-bold text-[#00ea8d]/70 group-hover:text-[#050505] opacity-60 group-hover:opacity-100">]</span>
        </button>
      </div>
    )
  }

  // Windows / Linux / Standard Controls (Far Right)
  return (
    <div
      style={{ '--wails-draggable': 'no-drag' } as React.CSSProperties}
      className="flex items-center h-full shrink-0 border-l border-[#2e2e2e] font-mono select-none"
    >
      <button
        type="button"
        onClick={() => WindowMinimise()}
        title="MINIMIZAR // [SYS_MIN]"
        aria-label="Minimizar ventana"
        className="h-full px-3 flex items-center justify-center gap-0.5 text-[#a7a49c] hover:text-[#050505] hover:bg-[#d9ff3f] border-r border-[#262626] transition-colors cursor-pointer group"
      >
        <span className="text-[10px] font-bold opacity-50 group-hover:opacity-100">[</span>
        <span className="font-black text-xs leading-none -translate-y-0.5">_</span>
        <span className="text-[10px] font-bold opacity-50 group-hover:opacity-100">]</span>
      </button>

      <button
        type="button"
        onClick={() => WindowToggleMaximise()}
        title="MAXIMIZAR / RESTAURAR // [SYS_EXPAND]"
        aria-label="Maximizar o restaurar ventana"
        className="h-full px-3 flex items-center justify-center gap-0.5 text-[#a7a49c] hover:text-[#050505] hover:bg-[#d9ff3f] border-r border-[#262626] transition-colors cursor-pointer group"
      >
        <span className="text-[10px] font-bold opacity-50 group-hover:opacity-100">[</span>
        <span className="font-black text-[11px] leading-none">⛶</span>
        <span className="text-[10px] font-bold opacity-50 group-hover:opacity-100">]</span>
      </button>

      <button
        type="button"
        onClick={() => Quit()}
        title="TERMINAR // [SYS_KILL]"
        aria-label="Cerrar ventana"
        className="h-full px-3.5 flex items-center justify-center gap-0.5 text-[#a7a49c] hover:text-[#050505] hover:bg-[#f87171] transition-colors cursor-pointer group"
      >
        <span className="text-[10px] font-bold opacity-50 group-hover:opacity-100">[</span>
        <span className="font-black text-xs leading-none">✕</span>
        <span className="text-[10px] font-bold opacity-50 group-hover:opacity-100">]</span>
      </button>
    </div>
  )
}
