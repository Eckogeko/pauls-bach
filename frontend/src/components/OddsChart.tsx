import { useEffect, useState, useCallback } from "react";
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  Tooltip,
  ResponsiveContainer,
} from "recharts";
import type { OddsSnapshot, OutcomeOdds } from "@/lib/types";
import { getOddsHistory } from "@/lib/api";
import { useEventStream } from "@/hooks/useEventStream";
import { Skeleton } from "@/components/ui/skeleton";

const COLORS = [
  "var(--primary)",
  "var(--chart-2)",
  "var(--chart-3)",
  "var(--chart-4)",
  "var(--chart-5)",
];

interface OddsChartProps {
  eventId: number;
  outcomes: OutcomeOdds[];
}

interface ChartPoint {
  time: string;
  [label: string]: string | number;
}

function formatTime(iso: string): string {
  const d = new Date(iso);
  return d.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" });
}

function buildChartData(
  snapshots: OddsSnapshot[],
  outcomes: OutcomeOdds[]
): ChartPoint[] {
  const labelMap = new Map<number, string>();
  for (const o of outcomes) {
    labelMap.set(o.outcome_id, o.label);
  }

  // Group snapshots by timestamp
  const groups = new Map<string, Map<number, number>>();
  for (const s of snapshots) {
    if (!groups.has(s.created_at)) {
      groups.set(s.created_at, new Map());
    }
    groups.get(s.created_at)!.set(s.outcome_id, s.odds);
  }

  const points: ChartPoint[] = [];
  for (const [ts, oddsMap] of groups) {
    const point: ChartPoint = { time: formatTime(ts) };
    for (const [outcomeId, odds] of oddsMap) {
      const label = labelMap.get(outcomeId) ?? `#${outcomeId}`;
      point[label] = Math.round(odds);
    }
    points.push(point);
  }

  return points;
}

export default function OddsChart({ eventId, outcomes }: OddsChartProps) {
  const [snapshots, setSnapshots] = useState<OddsSnapshot[] | null>(null);

  const fetchHistory = useCallback(() => {
    getOddsHistory(eventId)
      .then(setSnapshots)
      .catch(() => {});
  }, [eventId]);

  useEffect(() => {
    fetchHistory();
  }, [fetchHistory]);

  useEventStream(
    useCallback(
      (msg) => {
        if (
          msg.type === "odds_updated" &&
          (msg.data as { event_id: number }).event_id === eventId
        ) {
          fetchHistory();
        }
      },
      [eventId, fetchHistory]
    )
  );

  if (snapshots === null) {
    return <Skeleton className="h-[150px] w-full rounded-lg" />;
  }

  const data = buildChartData(snapshots, outcomes);

  if (data.length < 2) {
    return (
      <div className="flex h-[100px] items-center justify-center text-xs text-muted-foreground">
        Chart will appear after the first trade
      </div>
    );
  }

  const labels = outcomes.map((o) => o.label);

  return (
    <div className="h-[150px] w-full">
      <ResponsiveContainer width="100%" height="100%">
        <LineChart data={data} margin={{ top: 5, right: 5, bottom: 5, left: -20 }}>
          <XAxis
            dataKey="time"
            tick={{ fontSize: 10 }}
            stroke="var(--muted-foreground)"
            tickLine={false}
            axisLine={false}
          />
          <YAxis
            domain={[0, 100]}
            tick={{ fontSize: 10 }}
            stroke="var(--muted-foreground)"
            tickLine={false}
            axisLine={false}
            tickFormatter={(v: number) => `${v}%`}
          />
          <Tooltip
            contentStyle={{
              backgroundColor: "var(--popover)",
              borderColor: "var(--border)",
              borderRadius: "var(--radius)",
              fontSize: 12,
            }}
            formatter={(value: number | undefined) => [`${value ?? 0}%`]}
          />
          {labels.map((label, i) => (
            <Line
              key={label}
              type="monotone"
              dataKey={label}
              stroke={COLORS[i % COLORS.length]}
              strokeWidth={2}
              dot={false}
              activeDot={{ r: 3 }}
            />
          ))}
        </LineChart>
      </ResponsiveContainer>
    </div>
  );
}
