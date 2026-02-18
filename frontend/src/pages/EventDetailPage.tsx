import { useEffect, useState, useCallback } from "react";
import { useParams, useNavigate } from "react-router-dom";
import type { EventDetail, OutcomeOdds } from "@/lib/types";
import { getEvent, buyShares, sellShares } from "@/lib/api";
import { useAuth } from "@/context/AuthContext";
import { useEventStream } from "@/hooks/useEventStream";
import OddsChart from "@/components/OddsChart";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { toast } from "sonner";
import { Skeleton } from "@/components/ui/skeleton";
import {
  ArrowLeft,
  TrendingUp,
  TrendingDown,
  Loader2,
  Coins,
  Clock,
  CheckCircle2,
  CircleOff,
} from "lucide-react";

const statusConfig = {
  open: { label: "Open", variant: "default" as const, icon: TrendingUp },
  closed: { label: "Closed", variant: "secondary" as const, icon: Clock },
  resolved: {
    label: "Resolved",
    variant: "outline" as const,
    icon: CheckCircle2,
  },
};

export default function EventDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { updateBalance } = useAuth();
  const [event, setEvent] = useState<EventDetail | null>(null);
  const [loading, setLoading] = useState(true);
  const [tradeLoading, setTradeLoading] = useState(false);
  const [selectedOutcome, setSelectedOutcome] = useState<number | null>(null);
  const [tradeMode, setTradeMode] = useState<"buy" | "sell">("buy");
  const [amount, setAmount] = useState("");

  const fetchEvent = useCallback(() => {
    if (!id) return;
    getEvent(Number(id))
      .then(setEvent)
      .catch(() => toast.error("Failed to load event"))
      .finally(() => setLoading(false));
  }, [id]);

  useEffect(() => {
    fetchEvent();
  }, [fetchEvent]);

  useEventStream(
    useCallback(
      (msg) => {
        if (msg.type === "odds_updated") {
          const d = msg.data as { event_id: number; odds: OutcomeOdds[] };
          if (d.event_id === Number(id)) {
            setEvent((prev) => (prev ? { ...prev, odds: d.odds } : prev));
          }
        } else if (msg.type === "event_resolved") {
          const d = msg.data as { event_id: number };
          if (d.event_id === Number(id)) {
            fetchEvent();
          }
        } else if (msg.type === "connected") {
          fetchEvent();
        }
      },
      [id, fetchEvent]
    )
  );

  const handleTrade = async () => {
    if (!id || selectedOutcome === null || !amount) return;
    setTradeLoading(true);
    try {
      if (tradeMode === "buy") {
        const res = await buyShares(
          Number(id),
          selectedOutcome,
          Number(amount)
        );
        toast.success(res.message);
        updateOdds(res.odds);
        updateBalance(res.balance);
      } else {
        const res = await sellShares(
          Number(id),
          selectedOutcome,
          Math.floor(Number(amount))
        );
        toast.success(`${res.message} (+${res.points_back.toFixed(1)} pts)`);
        updateOdds(res.odds);
        updateBalance(res.balance);
      }
      setAmount("");
      setSelectedOutcome(null);
      fetchEvent();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Trade failed");
    } finally {
      setTradeLoading(false);
    }
  };

  const updateOdds = (newOdds: OutcomeOdds[]) => {
    setEvent((prev) => (prev ? { ...prev, odds: newOdds } : prev));
  };

  if (loading) {
    return (
      <div className="space-y-6">
        <Skeleton className="h-8 w-36" />
        <div className="space-y-2">
          <Skeleton className="h-8 w-2/3" />
          <Skeleton className="h-5 w-full" />
        </div>
        <div className="grid gap-4 md:grid-cols-3">
          <Card className="md:col-span-2">
            <CardContent className="space-y-3 pt-6">
              {Array.from({ length: 3 }).map((_, i) => (
                <Skeleton key={i} className="h-20 w-full rounded-lg" />
              ))}
            </CardContent>
          </Card>
          <Skeleton className="h-64 rounded-lg" />
        </div>
      </div>
    );
  }

  if (!event) {
    return (
      <div className="flex flex-col items-center justify-center py-16 text-center">
        <div className="mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-muted">
          <CircleOff className="h-8 w-8 text-muted-foreground" />
        </div>
        <p className="text-lg font-medium">Event not found</p>
        <p className="mt-1 text-sm text-muted-foreground">
          This event may have been removed or doesn't exist.
        </p>
        <Button
          variant="outline"
          className="mt-4 gap-1.5"
          onClick={() => navigate("/events")}
        >
          <ArrowLeft className="h-4 w-4" />
          Back to Markets
        </Button>
      </div>
    );
  }

  const config = statusConfig[event.status];
  const StatusIcon = config.icon;
  const totalOdds = event.odds.reduce((sum, o) => sum + o.odds, 0);
  const colors = [
    "bg-primary",
    "bg-chart-2",
    "bg-chart-3",
    "bg-chart-4",
    "bg-chart-5",
  ];

  return (
    <div className="animate-in fade-in duration-300 space-y-6">
      <Button
        variant="ghost"
        size="sm"
        onClick={() => navigate("/events")}
        className="gap-1.5"
      >
        <ArrowLeft className="h-4 w-4" />
        Back to Markets
      </Button>

      <div className="space-y-1">
        <div className="flex items-start gap-3">
          <h1 className="text-2xl font-bold tracking-tight">{event.title}</h1>
          <Badge variant={config.variant} className="mt-1 shrink-0 gap-1">
            <StatusIcon className="h-3 w-3" />
            {config.label}
          </Badge>
        </div>
        {event.description && (
          <p className="text-muted-foreground">{event.description}</p>
        )}
      </div>

      <div className="grid gap-4 md:grid-cols-3">
        {/* Odds Panel */}
        <Card className="md:col-span-2">
          <CardHeader>
            <CardTitle className="text-base">Outcomes</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <OddsChart eventId={event.id} outcomes={event.odds} />
            <div className="space-y-3">
            {event.odds.map((o, i) => {
              const pct =
                totalOdds > 0 ? Math.round((o.odds / totalOdds) * 100) : 0;
              const isWinner = event.winning_outcome_id === o.outcome_id;
              const isSelected = selectedOutcome === o.outcome_id;

              return (
                <button
                  key={o.outcome_id}
                  type="button"
                  onClick={() => {
                    if (event.status === "open") {
                      setSelectedOutcome(
                        isSelected ? null : o.outcome_id
                      );
                    }
                  }}
                  className={`w-full rounded-lg border p-3 text-left transition-all ${
                    isSelected
                      ? "border-primary bg-primary/5 ring-1 ring-primary"
                      : event.status === "open"
                        ? "hover:border-primary/50"
                        : ""
                  } ${isWinner ? "border-green-500 bg-green-50" : ""}`}
                  disabled={event.status !== "open"}
                >
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-2">
                      <div
                        className={`h-3 w-3 rounded-full ${colors[i % colors.length]}`}
                      />
                      <span className="font-medium">{o.label}</span>
                      {isWinner && (
                        <Badge
                          variant="outline"
                          className="border-green-500 text-green-700"
                        >
                          Winner
                        </Badge>
                      )}
                    </div>
                    <span className="text-lg font-bold">{pct}%</span>
                  </div>
                  <div className="mt-2 h-2 overflow-hidden rounded-full bg-secondary">
                    <div
                      className={`h-full ${colors[i % colors.length]} transition-all duration-500`}
                      style={{ width: `${pct}%` }}
                    />
                  </div>
                  <div className="mt-1 text-xs text-muted-foreground">
                    {Math.round(o.shares)} shares outstanding
                  </div>
                </button>
              );
            })}
            </div>
          </CardContent>
        </Card>

        {/* Trading Panel */}
        <div className="space-y-4">
          {event.status === "open" && (
            <Card>
              <CardHeader>
                <CardTitle className="text-base">Trade</CardTitle>
                <CardDescription>
                  {selectedOutcome
                    ? `Trading: ${event.odds.find((o) => o.outcome_id === selectedOutcome)?.label}`
                    : "Select an outcome to trade"}
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="flex gap-2">
                  <Button
                    variant={tradeMode === "buy" ? "default" : "outline"}
                    size="sm"
                    className="flex-1 gap-1"
                    onClick={() => setTradeMode("buy")}
                  >
                    <TrendingUp className="h-3.5 w-3.5" />
                    Buy
                  </Button>
                  <Button
                    variant={tradeMode === "sell" ? "default" : "outline"}
                    size="sm"
                    className="flex-1 gap-1"
                    onClick={() => setTradeMode("sell")}
                  >
                    <TrendingDown className="h-3.5 w-3.5" />
                    Sell
                  </Button>
                </div>

                {tradeMode === "sell" && (
                  <p className="text-xs text-amber-600 dark:text-amber-400">
                    Selling returns 50% of share value. The rest stays in the prize pool.
                  </p>
                )}

                <div className="space-y-2">
                  <Label>
                    {tradeMode === "buy" ? "Points to spend" : "Shares to sell"}
                  </Label>
                  <div className="relative">
                    <Input
                      type="number"
                      min="1"
                      step="1"
                      placeholder={tradeMode === "buy" ? "100" : "1"}
                      value={amount}
                      onChange={(e) => setAmount(e.target.value)}
                    />
                    <Coins className="absolute right-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                  </div>
                </div>

                <Button
                  className="w-full"
                  disabled={
                    !selectedOutcome || !amount || Number(amount) <= 0 || tradeLoading
                  }
                  onClick={handleTrade}
                >
                  {tradeLoading ? (
                    <Loader2 className="h-4 w-4 animate-spin" />
                  ) : tradeMode === "buy" ? (
                    "Buy Shares"
                  ) : (
                    "Sell Shares"
                  )}
                </Button>
              </CardContent>
            </Card>
          )}

          {/* User Positions */}
          {event.user_positions && event.user_positions.length > 0 && (
            <Card>
              <CardHeader>
                <CardTitle className="text-base">Your Positions</CardTitle>
              </CardHeader>
              <CardContent className="space-y-3">
                {event.user_positions.map((pos) => (
                  <div
                    key={pos.outcome_id}
                    className="flex items-center justify-between text-sm"
                  >
                    <div>
                      <div className="font-medium">{pos.outcome_label}</div>
                      <div className="text-xs text-muted-foreground">
                        Avg price: {pos.avg_price.toFixed(2)}
                      </div>
                    </div>
                    <div className="text-right">
                      <div className="font-medium">
                        {Math.round(pos.shares)} shares
                      </div>
                    </div>
                  </div>
                ))}
              </CardContent>
            </Card>
          )}
        </div>
      </div>
    </div>
  );
}
