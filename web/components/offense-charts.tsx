/**
 * Copyright 2025 The ChapaUY Authors
 * SPDX-License-Identifier: Apache-2.0
 */

"use client"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import {
  Bar,
  BarChart,
  CartesianGrid,
  Legend,
  Line,
  LineChart,
  XAxis,
  YAxis,
} from "recharts"
import {
  ChartContainer,
  ChartTooltip,
  type ChartConfig,
} from "@/components/ui/chart"
import { Skeleton } from "@/components/ui/skeleton"

interface OffenseChartsProps {
  dayOfWeekData: Record<string, Record<string, number>> | null
  dayOfYearData: Record<string, Record<string, number>> | null
  timeOfDayData: Record<string, Record<string, number>> | null
  isLoading: boolean
  groupBy: string
  onGroupByChange: (value: string) => void
}

function formatNumberFull(value: number): string {
  return new Intl.NumberFormat("es-UY").format(value)
}

function formatNumber(value: number): string {
  if (value >= 1000000) {
    return `${(value / 1000000).toFixed(1)}M`
  }
  if (value >= 1000) {
    return `${(value / 1000).toFixed(1)}K`
  }
  return value.toString()
}

const groupColors = [
  "hsl(200, 70%, 50%)", // blue
  "hsl(280, 60%, 50%)", // purple
  "hsl(30, 70%, 50%)", // orange
  "hsl(340, 60%, 50%)", // pink
  "hsl(160, 60%, 50%)", // teal
]

function transformChartData(
  data: Record<string, Record<string, number>> | null
) {
  if (!data) return { chartData: [], groups: [], config: {} }

  const groups = Object.keys(data).filter((key) => key !== "")
  const isGrouped =
    groups.length > 1 || (groups.length === 1 && groups[0] !== "")

  if (!isGrouped) {
    // No grouping - simple format
    const simpleData = data[""] || Object.values(data)[0] || {}
    const chartData = Object.entries(simpleData).map(([key, value]) => ({
      name: key,
      count: value,
    }))

    const config = {
      count: {
        label: "Ofensas",
        color: groupColors[0],
      },
    }

    return { chartData, groups: ["count"], config }
  }

  const allKeys = new Set<string>()
  groups.forEach((group) => {
    Object.keys(data[group] || {}).forEach((key) => allKeys.add(key))
  })

  const sortedKeys = Array.from(allKeys).sort()

  const chartData = sortedKeys.map((key) => {
    const row: any = { name: key }
    groups.forEach((group) => {
      row[group] = data[group]?.[key] ?? null
    })
    return row
  })

  const config: ChartConfig = {}
  groups.forEach((group, index) => {
    config[group] = {
      label: group,
      color: groupColors[index % groupColors.length],
    }
  })

  return { chartData, groups, config }
}

function ChartSkeleton() {
  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <Skeleton className="h-5 w-32" />
      </div>
      <div className="space-y-4">
        <Skeleton className="h-[300px] w-full" />
      </div>
    </div>
  )
}

function CustomTooltip({ active, payload }: any) {
  if (active && payload && payload.length) {
    return (
      <div className="bg-background rounded-lg border p-2 shadow-sm">
        <div className="grid gap-2">
          <span className="text-muted-foreground text-[0.70rem] uppercase">
            {payload[0].payload.name}
          </span>
          {payload.map((entry: any, index: number) => (
            <div key={index} className="flex items-center gap-2">
              <div
                className="h-2 w-2 rounded-full"
                style={{ backgroundColor: entry.color }}
              />
              <span className="text-muted-foreground text-xs">
                {entry.dataKey}:
              </span>
              <span className="text-foreground font-bold">
                {formatNumberFull(entry.value)}
              </span>
            </div>
          ))}
        </div>
      </div>
    )
  }
  return null
}

export function OffenseCharts({
  dayOfWeekData,
  dayOfYearData,
  timeOfDayData,
  isLoading,
  groupBy,
  onGroupByChange,
}: OffenseChartsProps) {
  const dayOfWeekTransformed = transformChartData(dayOfWeekData)
  const dayOfYearTransformed = transformChartData(dayOfYearData)
  const timeOfDayTransformed = transformChartData(timeOfDayData)

  if (isLoading) {
    return (
      <div className="space-y-4">
        <ChartSkeleton />
        <ChartSkeleton />
        <ChartSkeleton />
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h2 className="text-lg font-semibold">Gráficos de ofensas</h2>
        <div className="flex items-center gap-2">
          <span className="text-muted-foreground text-sm">Agrupar por:</span>
          <Select value={groupBy} onValueChange={onGroupByChange}>
            <SelectTrigger className="w-[180px]">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="none">Sin agrupar</SelectItem>
              <SelectItem value="year">Año</SelectItem>
            </SelectContent>
          </Select>
        </div>
      </div>

      <div className="space-y-6">
        {/* Day of Week Chart */}
        <div className="bg-card/50">
          <div className="flex items-center justify-between">
            <h3 className="text-lg font-semibold">Por día de la semana</h3>
          </div>
          <div className="space-y-4">
            <div className="h-[300px] min-h-[300px] w-full">
              <ChartContainer
                config={dayOfWeekTransformed.config}
                className="h-[300px] w-full"
              >
                <BarChart data={dayOfWeekTransformed.chartData}>
                  <CartesianGrid
                    strokeDasharray="3 3"
                    className="stroke-muted/30"
                  />
                  <XAxis
                    dataKey="name"
                    tickLine={false}
                    axisLine={false}
                    className="text-xs"
                  />
                  <YAxis
                    tickLine={false}
                    axisLine={false}
                    tickFormatter={formatNumber}
                  />
                  <ChartTooltip content={<CustomTooltip />} />
                  {dayOfWeekTransformed.groups.length > 1 && <Legend />}
                  {dayOfWeekTransformed.groups.map((group) => (
                    <Bar
                      key={group}
                      dataKey={group}
                      fill={`var(--color-${group})`}
                      stroke={`var(--color-${group})`}
                      strokeWidth={1}
                      radius={4}
                    />
                  ))}
                </BarChart>
              </ChartContainer>
            </div>
          </div>
        </div>

        {/* Day of Year Chart */}
        <div className="bg-card/50">
          <div className="flex items-center justify-between">
            <h3 className="text-lg font-semibold">Por día del año</h3>
          </div>
          <div className="space-y-4">
            <div className="h-[300px] min-h-[300px] w-full">
              <ChartContainer
                config={dayOfYearTransformed.config}
                className="h-[300px] w-full"
              >
                <LineChart data={dayOfYearTransformed.chartData}>
                  <CartesianGrid
                    strokeDasharray="3 3"
                    className="stroke-muted/30"
                  />
                  <XAxis
                    dataKey="name"
                    tickLine={false}
                    axisLine={false}
                    className="text-xs"
                    interval="preserveStartEnd"
                  />
                  <YAxis
                    tickLine={false}
                    axisLine={false}
                    tickFormatter={formatNumber}
                  />
                  <ChartTooltip content={<CustomTooltip />} />
                  {dayOfYearTransformed.groups.length > 1 && <Legend />}
                  {dayOfYearTransformed.groups.map((group) => (
                    <Line
                      key={group}
                      type="monotone"
                      dataKey={group}
                      stroke={`var(--color-${group})`}
                      strokeWidth={4}
                      dot={false}
                      connectNulls={true}
                    />
                  ))}
                </LineChart>
              </ChartContainer>
            </div>
          </div>
        </div>

        {/* Time of Day Chart */}
        <div className="bg-card/50">
          <div className="flex items-center justify-between">
            <h3 className="text-lg font-semibold">Por hora del día</h3>
          </div>
          <div className="space-y-4">
            <div className="h-[300px] min-h-[300px] w-full">
              <ChartContainer
                config={timeOfDayTransformed.config}
                className="h-[300px] w-full"
              >
                <BarChart data={timeOfDayTransformed.chartData}>
                  <CartesianGrid
                    strokeDasharray="3 3"
                    className="stroke-muted/30"
                  />
                  <XAxis
                    dataKey="name"
                    tickLine={false}
                    axisLine={false}
                    className="text-xs"
                  />
                  <YAxis
                    tickLine={false}
                    axisLine={false}
                    tickFormatter={formatNumber}
                  />
                  <ChartTooltip content={<CustomTooltip />} />
                  {timeOfDayTransformed.groups.length > 1 && <Legend />}
                  {timeOfDayTransformed.groups.map((group) => (
                    <Bar
                      key={group}
                      dataKey={group}
                      fill={`var(--color-${group})`}
                      stroke={`var(--color-${group})`}
                      strokeWidth={1}
                      radius={4}
                    />
                  ))}
                </BarChart>
              </ChartContainer>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
