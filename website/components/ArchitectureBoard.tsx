'use client'

import { Terminal, Globe, Boxes, ShieldCheck, Database, Layers, Server } from 'lucide-react'
import { CircuitBoard } from '@/components/ui/circuit-board'

// Orbita request/orchestration flow, drawn with the vendored CircuitBoard
// (componentry.dev) and tinted to the coral brand.
const CORAL_PULSE = 'rgba(244, 91, 72, 0.85)'
const CORAL_NODE = 'rgba(244, 91, 72, 0.7)'
const TRACE = 'rgba(148, 163, 184, 0.22)'
const GRID = 'rgba(244, 91, 72, 0.05)'

const ic = 'w-4 h-4'

const nodes = [
  { id: 'cli', x: 80, y: 90, label: 'grit CLI', icon: <Terminal className={ic} /> },
  { id: 'browser', x: 80, y: 250, label: 'Browser', icon: <Globe className={ic} /> },
  { id: 'traefik', x: 340, y: 50, label: 'Traefik', icon: <ShieldCheck className={ic} /> },
  { id: 'orbita', x: 340, y: 175, label: 'Orbita', icon: <Boxes className="h-5 w-5" />, size: 'lg' as const },
  { id: 'postgres', x: 590, y: 80, label: 'Postgres', icon: <Database className={ic} /> },
  { id: 'redis', x: 590, y: 175, label: 'Redis', icon: <Layers className={ic} /> },
  { id: 'swarm', x: 590, y: 270, label: 'Swarm', icon: <Server className={ic} /> },
]

const connections = [
  { from: 'cli', to: 'orbita', animated: true },
  { from: 'browser', to: 'orbita', animated: true },
  { from: 'orbita', to: 'traefik', bidirectional: true },
  { from: 'orbita', to: 'postgres', animated: true },
  { from: 'orbita', to: 'redis', animated: true },
  { from: 'orbita', to: 'swarm', animated: true },
]

export function ArchitectureBoard() {
  return (
    <div className="overflow-x-auto">
      <div className="mx-auto w-[680px] max-w-none px-4 py-4">
        <CircuitBoard
          variant="dark"
          nodes={nodes}
          connections={connections}
          width={680}
          height={340}
          pulseColor={CORAL_PULSE}
          nodeColor={CORAL_NODE}
          traceColor={TRACE}
          gridColor={GRID}
          pulseSpeed={2.2}
        />
      </div>
    </div>
  )
}
