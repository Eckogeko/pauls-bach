import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { getHistory } from "@/lib/api";
import { useAuth } from "@/context/AuthContext";
import type { HistoryEntry } from "@/lib/types";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { ScrollText } from "lucide-react";

const txTypeConfig = {
  buy: { label: "Buy", variant: "default" as const },
  sell: { label: "Sell", variant: "secondary" as const },
  payout: { label: "Payout", variant: "outline" as const },
  bonus: { label: "Bonus", variant: "outline" as const },
};

export default function HistoryPage() {
  const [history, setHistory] = useState<HistoryEntry[]>([]);
  const [loading, setLoading] = useState(true);
  const { user } = useAuth();

  useEffect(() => {
    if (!user) return;
    getHistory(user.id)
      .then(setHistory)
      .catch(() => {})
      .finally(() => setLoading(false));
  }, [user]);

  if (loading) {
    return (
      <div className="space-y-6">
        <Skeleton className="h-8 w-48" />
        <Card>
          <CardHeader>
            <Skeleton className="h-5 w-36" />
          </CardHeader>
          <CardContent className="space-y-2">
            {Array.from({ length: 5 }).map((_, i) => (
              <Skeleton key={i} className="h-14 w-full rounded-lg" />
            ))}
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="animate-in fade-in duration-300 space-y-6">
      <h1 className="text-2xl font-bold tracking-tight">Trade History</h1>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">Your Transactions</CardTitle>
        </CardHeader>
        <CardContent>
          {history.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-12 text-center">
              <div className="mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-muted">
                <ScrollText className="h-8 w-8 text-muted-foreground" />
              </div>
              <p className="text-lg font-medium">No trades yet</p>
              <p className="mt-1 text-sm text-muted-foreground">
                Your trading history will appear here after your first bet.
              </p>
            </div>
          ) : (
            <div className="space-y-2">
              {history.map((entry) => {
                const config = txTypeConfig[entry.tx_type];
                const isPositive =
                  entry.tx_type === "sell" ||
                  entry.tx_type === "payout" ||
                  entry.tx_type === "bonus";

                return (
                  <div
                    key={entry.id}
                    className="flex items-center justify-between rounded-lg border px-3 py-2.5"
                  >
                    <div className="min-w-0 flex-1">
                      <div className="flex items-center gap-2">
                        <Link
                          to={`/events/${entry.event_id}`}
                          className="truncate font-medium hover:underline"
                        >
                          {entry.event_title}
                        </Link>
                        <Badge variant={config.variant} className="shrink-0">
                          {config.label}
                        </Badge>
                      </div>
                      <div className="mt-0.5 text-xs text-muted-foreground">
                        {entry.outcome_label} &middot;{" "}
                        {entry.shares.toFixed(1)} shares &middot;{" "}
                        {new Date(entry.created_at).toLocaleDateString()}
                      </div>
                    </div>
                    <span
                      className={`ml-3 shrink-0 font-semibold tabular-nums ${
                        isPositive ? "text-green-600" : "text-red-500"
                      }`}
                    >
                      {isPositive ? "+" : "-"}
                      {Math.abs(entry.points).toLocaleString()} pts
                    </span>
                  </div>
                );
              })}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
