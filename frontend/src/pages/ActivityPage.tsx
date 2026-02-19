import { useEffect, useState, useCallback } from "react";
import { Link } from "react-router-dom";
import { getActivity } from "@/lib/api";
import { useEventStream } from "@/hooks/useEventStream";
import type { ActivityEntry } from "@/lib/types";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Rss,
  TrendingUp,
  Gavel,
  Trophy,
  Grid3X3,
  Award,
} from "lucide-react";

const typeIcons: Record<string, typeof Rss> = {
  trade: TrendingUp,
  event_created: Rss,
  event_resolved: Gavel,
  payout: Trophy,
  bingo_resolved: Grid3X3,
  bingo_winner: Award,
};

function timeAgo(dateStr: string): string {
  const now = Date.now();
  const then = new Date(dateStr).getTime();
  const diff = Math.max(0, now - then);
  const seconds = Math.floor(diff / 1000);
  if (seconds < 60) return "just now";
  const minutes = Math.floor(seconds / 60);
  if (minutes < 60) return `${minutes}m ago`;
  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `${hours}h ago`;
  const days = Math.floor(hours / 24);
  return `${days}d ago`;
}

export default function ActivityPage() {
  const [entries, setEntries] = useState<ActivityEntry[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    getActivity()
      .then(setEntries)
      .catch(() => {})
      .finally(() => setLoading(false));
  }, []);

  useEventStream(
    useCallback((msg) => {
      if (msg.type === "activity_new" && msg.data) {
        const entry = msg.data as unknown as ActivityEntry;
        setEntries((prev) => [entry, ...prev].slice(0, 50));
      }
      if (msg.type === "connected") {
        getActivity().then(setEntries).catch(() => {});
      }
    }, [])
  );

  if (loading) {
    return (
      <div className="space-y-3">
        {Array.from({ length: 5 }).map((_, i) => (
          <Skeleton key={i} className="h-14 w-full rounded-lg" />
        ))}
      </div>
    );
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Rss className="h-5 w-5" />
          Activity Feed
        </CardTitle>
      </CardHeader>
      <CardContent>
        {entries.length === 0 ? (
          <p className="text-center text-muted-foreground py-8">
            No activity yet — trades, resolutions, and bingo wins will appear here.
          </p>
        ) : (
          <div className="space-y-1">
            {entries.map((entry) => {
              const Icon = typeIcons[entry.type] || Rss;
              const content = (
                <div className="flex items-start gap-3 rounded-lg px-3 py-2.5 hover:bg-muted/50 transition-colors">
                  <Icon className="h-4 w-4 mt-0.5 shrink-0 text-muted-foreground" />
                  <span className="text-sm flex-1">{entry.message}</span>
                  <span className="text-xs text-muted-foreground whitespace-nowrap">
                    {timeAgo(entry.created_at)}
                  </span>
                </div>
              );

              if (entry.event_id > 0) {
                return (
                  <Link key={entry.id} to={`/events/${entry.event_id}`}>
                    {content}
                  </Link>
                );
              }
              return <div key={entry.id}>{content}</div>;
            })}
          </div>
        )}
      </CardContent>
    </Card>
  );
}
