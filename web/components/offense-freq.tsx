/**
 * Copyright 2025 The ChapaUY Authors
 * SPDX-License-Identifier: Apache-2.0
 */

"use client"

import React, { useMemo } from "react"
import { FreqData } from "@/lib/types"
import { Card } from "@/components/ui/card"
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip"
import { cn } from "@/lib/utils"

interface OffenseFreqProps {
  data: FreqData | null
  isLoading?: boolean
  isYearFiltered?: boolean
}

const DAYS_ES = ["Domingo", "Lunes", "Martes", "Miércoles", "Jueves", "Viernes", "Sábado"]
const MONTHS_ES = [
  "Enero", "Febrero", "Marzo", "Abril", "Mayo", "Junio", 
  "Julio", "Agosto", "Septiembre", "Octubre", "Noviembre", "Diciembre"
]

function formatCount(val: number): string {
  return new Intl.NumberFormat("es-UY").format(val)
}

function getLocalDateStr(date: Date) {
  const y = date.getFullYear()
  const m = String(date.getMonth() + 1).padStart(2, '0')
  const d = String(date.getDate()).padStart(2, '0')
  return `${y}-${m}-${d}`
}

function getLocalMax(dataset: Record<string, Record<string, number>>, limitYears?: string[]) {
  let max = 0
  Object.entries(dataset).forEach(([year, yearData]) => {
    if (limitYears && !limitYears.includes(year)) return
    Object.values(yearData).forEach(val => {
      if (val > max) max = val
    })
  })
  return max || 1
}

function getIntensityColor(v: number | undefined, m: number) {
  if (!v || v === 0) {
    return "bg-zinc-100/50 dark:bg-white/[0.03] border border-zinc-200/50 dark:border-white/[0.05]"
  }
  const ratio = v / m
  if (ratio < 0.2) return "bg-orange-200 dark:bg-orange-900/40 border-transparent"
  if (ratio < 0.45) return "bg-orange-400 dark:bg-orange-700/60 border-transparent"
  if (ratio < 0.7) return "bg-orange-600 dark:bg-orange-500/80 border-transparent"
  if (ratio < 0.9) return "bg-orange-800 dark:bg-orange-400 border-transparent"
  return "bg-orange-950 dark:bg-orange-300 border-transparent"
}

function ColorBox({ val, max, label, size = "h-[14px] w-[14px]" }: { val: number, max: number, label: string, size?: string }) {
  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <div className={cn("rounded-[2px] transition-colors cursor-default shrink-0", size, getIntensityColor(val, max))} />
      </TooltipTrigger>
      <TooltipContent>
        <p className="text-xs">
          {label}: <span className="font-bold">{formatCount(val)}</span>
        </p>
      </TooltipContent>
    </Tooltip>
  )
}

export function OffenseFreq({ data, isLoading, isYearFiltered }: OffenseFreqProps) {
  if (isLoading) {
    return (
      <div className="flex h-96 w-full items-center justify-center">
        <p className="text-muted-foreground">Cargando frecuencia...</p>
      </div>
    )
  }

  if (!data) return null

  const currentYear = new Date().getFullYear()

  const years = useMemo(() => {
    const allYears = new Set<string>()
    Object.keys(data.dow).forEach(y => allYears.add(y))
    Object.keys(data.month).forEach(y => allYears.add(y))
    Object.keys(data.hour).forEach(y => allYears.add(y))
    let sorted = Array.from(allYears).sort().reverse()
    if (!isYearFiltered) {
      sorted = sorted.filter(y => parseInt(y) <= currentYear)
      if (sorted.length > 4) return sorted.slice(0, 4)
    }
    return sorted
  }, [data, isYearFiltered, currentYear])

  const maxDow = useMemo(() => getLocalMax(data.dow, years), [data.dow, years])
  const maxMonth = useMemo(() => getLocalMax(data.month, years), [data.month, years])
  const maxHour = useMemo(() => getLocalMax(data.hour, years), [data.hour, years])
  
  // Daily max excluding future unless filtered
  const rollingMaxDaily = useMemo(() => {
    let max = 0
    Object.entries(data.daily).forEach(([year, yd]) => {
      if (!isYearFiltered && parseInt(year) > currentYear) return
      Object.values(yd).forEach(v => { if (v > max) max = v })
    })
    return max || 1
  }, [data.daily, isYearFiltered, currentYear])

  return (
    <TooltipProvider delayDuration={0}>
      <div className="space-y-8 py-4">
        <div className="flex flex-wrap gap-6 items-start">
          {/* Day of Week Matrix */}
          <Card className="p-4 w-fit min-w-0 border-muted/20 shadow-none">
            <h3 className="mb-6 text-sm font-medium text-muted-foreground uppercase tracking-wider text-sm">Día</h3>
            <div className="space-y-2">
              <div className="flex gap-1.5 text-[11px] text-muted-foreground">
                <div className="w-12 shrink-0" />
                <div className="flex gap-1.5">
                  {DAYS_ES.map((d) => (
                    <span key={d} className="w-[14px] text-center">{d[0]}</span>
                  ))}
                </div>
              </div>
              {years.map(year => (
                <div key={year} className="flex items-center gap-1.5">
                  <span className="w-12 shrink-0 text-sm font-medium text-muted-foreground/70">{year}</span>
                  <div className="flex gap-1.5">
                    {Array.from({ length: 7 }).map((_, i) => (
                      <ColorBox key={i} val={data.dow[year]?.[i.toString()] || 0} max={maxDow} label={`${year}, ${DAYS_ES[i]}`} />
                    ))}
                  </div>
                </div>
              ))}
            </div>
          </Card>

          {/* Month Matrix */}
          <Card className="p-4 w-fit min-w-0 border-muted/20 shadow-none">
            <h3 className="mb-6 text-sm font-medium text-muted-foreground uppercase tracking-wider">Mes</h3>
            <div className="space-y-2">
              <div className="flex gap-1.5 text-[11px] text-muted-foreground">
                <div className="w-12 shrink-0" />
                <div className="flex gap-1.5">
                  {MONTHS_ES.map((m) => (
                    <span key={m} className="w-[14px] text-center text-[9px]">{m[0]}</span>
                  ))}
                </div>
              </div>
              {years.map(year => (
                <div key={year} className="flex items-center gap-1.5">
                  <span className="w-12 shrink-0 text-sm font-medium text-muted-foreground/70">{year}</span>
                  <div className="flex gap-1.5">
                    {Array.from({ length: 12 }).map((_, i) => {
                      const m = (i + 1).toString().padStart(2, '0')
                      return (
                        <ColorBox key={i} val={data.month[year]?.[m] || 0} max={maxMonth} label={`${year}, ${MONTHS_ES[i]}`} />
                      )
                    })}
                  </div>
                </div>
              ))}
            </div>
          </Card>

          {/* Hour Matrix */}
          <Card className="p-4 w-fit min-w-0 overflow-hidden border-muted/20 shadow-none">
            <h3 className="mb-6 text-sm font-medium text-muted-foreground uppercase tracking-wider">Hora (24h)</h3>
            <div className="overflow-x-auto pb-2">
              <div className="space-y-2 min-w-max">
                <div className="flex gap-1 text-[11px] text-muted-foreground">
                  <div className="w-12 shrink-0" />
                  <div className="flex justify-between w-[430px] px-0.5">
                    <span>0h</span><span>6h</span><span>12h</span><span>18h</span><span>23h</span>
                  </div>
                </div>
                {years.map(year => (
                  <div key={year} className="flex items-center gap-1">
                    <span className="w-12 shrink-0 text-sm font-medium text-muted-foreground/70">{year}</span>
                    <div className="flex gap-1">
                      {Array.from({ length: 24 }).map((_, i) => {
                        const h = i.toString().padStart(2, '0')
                        return (
                          <ColorBox key={i} val={data.hour[year]?.[h] || 0} max={maxHour} label={`${year}, ${i}:00hs`} />
                        )
                      })}
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </Card>
        </div>

        {/* Sliding Heatmap */}
        <Card className="p-4 border-muted/20 shadow-none">
          <h3 className="mb-6 text-sm font-medium text-muted-foreground uppercase tracking-wider">
            {isYearFiltered ? "Actividad en el período seleccionado" : "Actividad en los últimos 12 meses"}
          </h3>
          <div className="overflow-x-auto pb-4 pt-8">
            <SlidingHeatmap data={data.daily} max={rollingMaxDaily} isYearFiltered={isYearFiltered} currentYear={currentYear} />
          </div>
        </Card>
      </div>
    </TooltipProvider>
  )
}

function SlidingHeatmap({ data, max, isYearFiltered, currentYear }: { data: Record<string, Record<string, number>>, max: number, isYearFiltered?: boolean, currentYear: number }) {
  const range = useMemo(() => {
    let latest = new Date()
    let hasData = false
    Object.keys(data).forEach(year => {
      // Ignore future years unless explicitly filtered
      if (!isYearFiltered && parseInt(year) > currentYear) return

      Object.keys(data[year]).forEach(dateStr => {
        const d = new Date(dateStr + "T12:00:00") 
        if (!hasData || d > latest) {
          // Double check: if not filtered, latest cannot exceed today
          if (!isYearFiltered && d > new Date()) return
          latest = d
          hasData = true
        }
      })
    })

    const end = new Date(latest)
    end.setDate(end.getDate() + (6 - end.getDay()))
    const start = new Date(end)
    start.setFullYear(start.getFullYear() - 1)
    start.setDate(start.getDate() - start.getDay())
    const weeks = []
    let current = new Date(start)
    while (current <= end) {
      const week = []
      for (let i = 0; i < 7; i++) {
        week.push(new Date(current))
        current.setDate(current.getDate() + 1)
      }
      weeks.push(week)
    }
    return weeks
  }, [data, isYearFiltered, currentYear])

  return (
    <div className="flex gap-2 min-w-max">
      <div className="flex flex-col justify-between pt-1 text-[11px] text-muted-foreground h-[124px]">
        <span className="h-[14px]">Dom</span><span className="h-[14px]">Mar</span><span className="h-[14px]">Jue</span><span className="h-[14px]">Sáb</span>
      </div>
      <div className="flex-1">
        <div className="flex gap-1">
          {range.map((week, weekIdx) => {
            let monthLabel: string | null = null
            for (let r = 0; r < 7; r++) { if (week[r].getDate() === 1) { monthLabel = MONTHS_ES[week[r].getMonth()].slice(0, 3); break } }
            return (
              <div key={weekIdx} className="flex flex-col gap-1 relative">
                {monthLabel && <span className="absolute -top-7 left-0 text-[11px] text-muted-foreground whitespace-nowrap font-bold">{monthLabel}</span>}
                {week.map((day, rowIdx) => {
                  const dateKey = getLocalDateStr(day)
                  const val = data[day.getFullYear().toString()]?.[dateKey] || 0
                  return (
                    <Tooltip key={rowIdx}>
                      <TooltipTrigger asChild>
                        <div className={cn("h-[14px] w-[14px] rounded-[2px] transition-colors cursor-default", getIntensityColor(val, max))} />
                      </TooltipTrigger>
                      <TooltipContent>
                        <p className="text-xs">{day.toLocaleDateString('es-UY', { day: 'numeric', month: 'long', year: 'numeric' })}: <span className="font-bold">{formatCount(val)}</span></p>
                      </TooltipContent>
                    </Tooltip>
                  )
                })}
              </div>
            )
          })}
        </div>
      </div>
    </div>
  )
}
