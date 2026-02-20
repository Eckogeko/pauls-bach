import { useEffect, useState, useCallback } from "react";
import { Link } from "react-router-dom";
import type { Event, OutcomeOdds } from "@/lib/types";
import { getEvents, createUserEvent } from "@/lib/api";
import { useEventStream } from "@/hooks/useEventStream";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Skeleton } from "@/components/ui/skeleton";
import { toast } from "sonner";
import { TrendingUp, Clock, CheckCircle2, BarChart3, Plus, X, Loader2 } from "lucide-react";

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

function OddsBar({ odds, bettors }: { odds: Event["odds"]; bettors?: Record<number, string[]> }) {
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
        {odds.map((o, i) => {
          const names = bettors?.[o.outcome_id] ?? [];
          return (
            <div key={o.outcome_id} className="flex items-center gap-1.5 text-xs">
              <div
                className={`h-2 w-2 rounded-full ${colors[i % colors.length]}`}
              />
              <span className="text-muted-foreground">{o.label}</span>
              <span className="font-medium">
                {Math.round((o.odds / total) * 100)}%
              </span>
              {names.length > 0 && (
                <span className="text-muted-foreground">
                  · {names.join(", ")}
                </span>
              )}
            </div>
          );
        })}
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

function ResolvedCarousel({ events }: { events: Event[] }) {
  const resolved = events
    .filter((e) => e.status === "resolved")
    .sort((a, b) => (b.resolved_at ?? "").localeCompare(a.resolved_at ?? ""))
    .slice(0, 5);

  if (resolved.length === 0) return null;

  const items = resolved.length >= 2 ? [...resolved, ...resolved] : resolved;
  const shouldAnimate = resolved.length >= 2;

  return (
    <div className="relative -mx-4 px-4 overflow-hidden">
      <div
        className={`flex gap-3 w-max ${shouldAnimate ? "animate-marquee hover:[animation-play-state:paused]" : ""}`}
      >
        {items.map((event, i) => {
          const winner = event.odds.find(
            (o) => o.outcome_id === event.winning_outcome_id
          );
          return (
            <Link key={`${event.id}-${i}`} to={`/events/${event.id}`} className="shrink-0 w-[200px]">
              <div className="rounded-lg border bg-card px-3 py-2.5 hover:bg-accent/50 transition-colors h-full">
                <p className="text-xs font-medium leading-snug line-clamp-2">{event.title}</p>
                {winner && (
                  <div className="mt-1.5 flex items-center gap-1">
                    <CheckCircle2 className={`h-3 w-3 shrink-0 ${winner.label === "No" ? "text-red-600" : "text-green-600"}`} />
                    <span className={`text-[11px] font-medium truncate ${winner.label === "No" ? "text-red-700 dark:text-red-400" : "text-green-700 dark:text-green-400"}`}>{winner.label}</span>
                  </div>
                )}
              </div>
            </Link>
          );
        })}
      </div>
    </div>
  );
}

export default function EventsPage() {
  const [events, setEvents] = useState<Event[]>([]);
  const [loading, setLoading] = useState(true);
  const [filter, setFilter] = useState<StatusFilter>("open");

  // Create bet dialog state
  const [createOpen, setCreateOpen] = useState(false);
  const [createLoading, setCreateLoading] = useState(false);
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [eventType, setEventType] = useState("binary");
  const [outcomes, setOutcomes] = useState(["Yes", "No"]);
  const [showAdvanced, setShowAdvanced] = useState(false);

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

  const handleEventTypeChange = (value: string) => {
    setEventType(value);
    if (value === "binary") {
      setOutcomes(["Yes", "No"]);
    } else {
      setOutcomes(["", "", ""]);
    }
  };

  const handleCreate = async () => {
    if (!title.trim()) {
      toast.error("Title is required");
      return;
    }
    if (outcomes.some((o) => !o.trim())) {
      toast.error("All outcomes must have labels");
      return;
    }
    setCreateLoading(true);
    try {
      await createUserEvent(title.trim(), description.trim(), eventType, outcomes);
      toast.success("Bet created");
      setTitle("");
      setDescription("");
      setEventType("binary");
      setOutcomes(["Yes", "No"]);
      setCreateOpen(false);
      setShowAdvanced(false);
      fetchEvents();
    } catch (err: unknown) {
      toast.error(err instanceof Error ? err.message : "Failed to create bet");
    } finally {
      setCreateLoading(false);
    }
  };

  const filtered =
    filter === "all" ? events : events.filter((e) => e.status === filter);

  return (
    <div className="animate-in fade-in duration-300 space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold tracking-tight">Markets</h1>
        <div className="flex items-center gap-3">
          <h2 className="text-xl font-medium text-muted-foreground hidden sm:block">Recent Events</h2>
          <Dialog open={createOpen} onOpenChange={setCreateOpen}>
            <DialogTrigger asChild>
              <Button size="sm" className="gap-1.5">
                <Plus className="h-4 w-4" />
                Create Bet
              </Button>
            </DialogTrigger>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>Create Bet</DialogTitle>
                <DialogDescription>
                  Defaults to Yes/No. Earn 10 pts if 5+ players bet on it!
                </DialogDescription>
              </DialogHeader>
              <div className="space-y-4">
                <div className="space-y-2">
                  <Label>What's the bet?</Label>
                  <Input
                    placeholder="Will Paul finish his drink?"
                    value={title}
                    onChange={(e) => setTitle(e.target.value)}
                    autoFocus
                  />
                </div>
                <div className="space-y-2">
                  <Label>Details (optional)</Label>
                  <Input
                    placeholder="Additional context..."
                    value={description}
                    onChange={(e) => setDescription(e.target.value)}
                  />
                </div>
                <button
                  type="button"
                  className="text-xs text-muted-foreground hover:text-foreground transition-colors"
                  onClick={() => {
                    setShowAdvanced(!showAdvanced);
                    if (showAdvanced) {
                      setEventType("binary");
                      setOutcomes(["Yes", "No"]);
                    }
                  }}
                >
                  {showAdvanced ? "- Hide options" : "+ Multiple outcomes"}
                </button>
                {showAdvanced && (
                  <div className="space-y-3 rounded-lg border p-3">
                    <div className="space-y-2">
                      <Label className="text-xs">Type</Label>
                      <Select value={eventType} onValueChange={handleEventTypeChange}>
                        <SelectTrigger>
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="binary">Binary (Yes/No)</SelectItem>
                          <SelectItem value="multi">Multiple Outcomes</SelectItem>
                        </SelectContent>
                      </Select>
                    </div>
                    {eventType === "multi" && (
                      <div className="space-y-2">
                        <Label className="text-xs">Outcomes</Label>
                        {outcomes.map((outcome, i) => (
                          <div key={i} className="flex gap-2">
                            <Input
                              placeholder={`Outcome ${i + 1}`}
                              value={outcome}
                              onChange={(e) => {
                                const next = [...outcomes];
                                next[i] = e.target.value;
                                setOutcomes(next);
                              }}
                            />
                            {outcomes.length > 2 && (
                              <Button
                                variant="ghost"
                                size="icon"
                                onClick={() =>
                                  setOutcomes(outcomes.filter((_, j) => j !== i))
                                }
                              >
                                <X className="h-4 w-4" />
                              </Button>
                            )}
                          </div>
                        ))}
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => setOutcomes([...outcomes, ""])}
                        >
                          <Plus className="mr-1 h-3.5 w-3.5" />
                          Add Outcome
                        </Button>
                      </div>
                    )}
                  </div>
                )}
                <Button
                  className="w-full"
                  onClick={handleCreate}
                  disabled={createLoading}
                >
                  {createLoading ? (
                    <Loader2 className="h-4 w-4 animate-spin" />
                  ) : (
                    "Create Bet"
                  )}
                </Button>
              </div>
            </DialogContent>
          </Dialog>
        </div>
      </div>

      {/* Recent results carousel */}
      <ResolvedCarousel events={events} />

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
              ? "No markets yet. Create the first bet!"
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
                    {event.status === "resolved" && event.winning_outcome_id ? (
                      <div className="flex items-center gap-1.5">
                        {(() => {
                          const winner = event.odds.find((o) => o.outcome_id === event.winning_outcome_id);
                          if (!winner) return null;
                          const isNo = winner.label === "No";
                          return (
                            <>
                              <CheckCircle2 className={`h-4 w-4 shrink-0 ${isNo ? "text-red-600" : "text-green-600"}`} />
                              <span className={`text-sm font-medium ${isNo ? "text-red-700 dark:text-red-400" : "text-green-700 dark:text-green-400"}`}>
                                {winner.label}
                              </span>
                            </>
                          );
                        })()}
                      </div>
                    ) : (
                      <OddsBar odds={event.odds} bettors={event.bettors} />
                    )}
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
