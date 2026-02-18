import { useEffect, useState, useCallback } from "react";
import { Link } from "react-router-dom";
import type { Event, OutcomeOdds } from "@/lib/types";
import { getEvents } from "@/lib/api";
import { useEventStream } from "@/hooks/useEventStream";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Skeleton } from "@/components/ui/skeleton";
import { TrendingUp, Clock, CheckCircle2, BarChart3 } from "lucide-react";

type StatusFilter = "all" | "open" | "closed" | "resolved";

const statusConfig = {
  open: { label: "Open", variant: "default" as const, icon: TrendingUp },
  closed: { label: "Closed", variant: "secondary" as const, icon: Clock },
  resolved: {
    label: "Resolved",
    variant: "outline" as const,
    icon: CheckCircle2,
  },
};

function OddsBar({ odds }: { odds: Event["odds"] }) {
  const total = odds.reduce((sum, o) => sum + o.odds, 0);
  if (total === 0) return null;

  const colors = [
    "bg-primary",
    "bg-chart-2",
    "bg-chart-3",
    "bg-chart-4",
    "bg-chart-5",
  ];

  return (
    <div className="space-y-2">
      <div className="flex h-2.5 overflow-hidden rounded-full bg-secondary">
        {odds.map((o, i) => (
          <div
            key={o.outcome_id}
            className={`${colors[i % colors.length]} transition-all duration-500`}
            style={{ width: `${(o.odds / total) * 100}%` }}
          />
        ))}
      </div>
      <div className="flex flex-wrap gap-x-4 gap-y-1">
        {odds.map((o, i) => (
          <div key={o.outcome_id} className="flex items-center gap-1.5 text-xs">
            <div
              className={`h-2 w-2 rounded-full ${colors[i % colors.length]}`}
            />
            <span className="text-muted-foreground">{o.label}</span>
            <span className="font-medium">
              {Math.round((o.odds / total) * 100)}%
            </span>
          </div>
        ))}
      </div>
    </div>
  );
}

function EventCardSkeleton() {
  return (
    <Card>
      <CardHeader className="pb-3">
        <div className="flex items-start justify-between gap-2">
          <Skeleton className="h-5 w-3/4" />
          <Skeleton className="h-5 w-16" />
        </div>
        <Skeleton className="mt-2 h-4 w-full" />
      </CardHeader>
      <CardContent>
        <Skeleton className="h-2.5 w-full rounded-full" />
        <div className="mt-2 flex gap-4">
          <Skeleton className="h-3 w-16" />
          <Skeleton className="h-3 w-16" />
        </div>
      </CardContent>
    </Card>
  );
}

export default function EventsPage() {
  const [events, setEvents] = useState<Event[]>([]);
  const [loading, setLoading] = useState(true);
  const [filter, setFilter] = useState<StatusFilter>("open");

  const fetchEvents = useCallback(() => {
    getEvents()
      .then(setEvents)
      .catch(() => {})
      .finally(() => setLoading(false));
  }, []);

  useEffect(() => {
    fetchEvents();
  }, [fetchEvents]);

  useEventStream(
    useCallback(
      (msg) => {
        if (msg.type === "odds_updated") {
          const { event_id, odds } = msg.data as {
            event_id: number;
            odds: OutcomeOdds[];
          };
          setEvents((prev) =>
            prev.map((e) => (e.id === event_id ? { ...e, odds } : e))
          );
        } else if (
          msg.type === "event_created" ||
          msg.type === "event_resolved" ||
          msg.type === "connected"
        ) {
          fetchEvents();
        }
      },
      [fetchEvents]
    )
  );

  const filtered =
    filter === "all" ? events : events.filter((e) => e.status === filter);

  return (
    <div className="animate-in fade-in duration-300 space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold tracking-tight">Markets</h1>
      </div>

      <Tabs
        value={filter}
        onValueChange={(v) => setFilter(v as StatusFilter)}
      >
        <TabsList>
          <TabsTrigger value="all">All</TabsTrigger>
          <TabsTrigger value="open">Open</TabsTrigger>
          <TabsTrigger value="closed">Closed</TabsTrigger>
          <TabsTrigger value="resolved">Resolved</TabsTrigger>
        </TabsList>
      </Tabs>

      {loading ? (
        <div className="grid gap-4 sm:grid-cols-2">
          {Array.from({ length: 4 }).map((_, i) => (
            <EventCardSkeleton key={i} />
          ))}
        </div>
      ) : filtered.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-16 text-center">
          <div className="mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-muted">
            <BarChart3 className="h-8 w-8 text-muted-foreground" />
          </div>
          <p className="text-lg font-medium">No markets found</p>
          <p className="mt-1 text-sm text-muted-foreground">
            {filter === "all"
              ? "Markets will appear here once an admin creates them."
              : `No ${filter} markets right now. Try a different filter.`}
          </p>
        </div>
      ) : (
        <div className="grid gap-4 sm:grid-cols-2">
          {filtered.map((event) => {
            const config = statusConfig[event.status];
            const StatusIcon = config.icon;
            return (
              <Link key={event.id} to={`/events/${event.id}`}>
                <Card className="h-full transition-all duration-200 hover:shadow-md hover:-translate-y-0.5">
                  <CardHeader className="pb-3">
                    <div className="flex items-start justify-between gap-2">
                      <CardTitle className="text-base leading-snug">
                        {event.title}
                      </CardTitle>
                      <Badge variant={config.variant} className="shrink-0 gap-1">
                        <StatusIcon className="h-3 w-3" />
                        {config.label}
                      </Badge>
                    </div>
                    {event.description && (
                      <p className="text-sm text-muted-foreground line-clamp-2">
                        {event.description}
                      </p>
                    )}
                  </CardHeader>
                  <CardContent>
                    <OddsBar odds={event.odds} />
                  </CardContent>
                </Card>
              </Link>
            );
          })}
        </div>
      )}
    </div>
  );
}
