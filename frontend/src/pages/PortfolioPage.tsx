import { useEffect, useState, useCallback } from "react";
import { Link } from "react-router-dom";
import { getPortfolio } from "@/lib/api";
import { useEventStream } from "@/hooks/useEventStream";
import type { Portfolio } from "@/lib/types";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Briefcase,
  Coins,
  TrendingUp,
  BarChart3,
} from "lucide-react";

export default function PortfolioPage() {
  const [portfolio, setPortfolio] = useState<Portfolio | null>(null);
  const [loading, setLoading] = useState(true);

  const fetchPortfolio = useCallback(() => {
    getPortfolio()
      .then(setPortfolio)
      .catch(() => {})
      .finally(() => setLoading(false));
  }, []);

  useEffect(() => {
    fetchPortfolio();
  }, [fetchPortfolio]);

  useEventStream(
    useCallback(
      (msg) => {
        if (
          msg.type === "odds_updated" ||
          msg.type === "event_resolved" ||
          msg.type === "connected"
        ) {
          fetchPortfolio();
        }
      },
      [fetchPortfolio]
    )
  );

  if (loading) {
    return (
      <div className="space-y-6">
        <Skeleton className="h-8 w-48" />
        <div className="grid gap-4 sm:grid-cols-3">
          {Array.from({ length: 3 }).map((_, i) => (
            <Skeleton key={i} className="h-24 rounded-lg" />
          ))}
        </div>
        <Card>
          <CardContent className="space-y-2 pt-6">
            {Array.from({ length: 4 }).map((_, i) => (
              <Skeleton key={i} className="h-16 w-full rounded-lg" />
            ))}
          </CardContent>
        </Card>
      </div>
    );
  }

  const positions = portfolio?.positions ?? [];
  const totalInvested = portfolio?.total_invested ?? 0;
  const totalPotential = portfolio?.total_potential ?? 0;
  const activeMarkets = portfolio?.active_markets ?? 0;
  const netProfit = totalPotential - totalInvested;

  return (
    <div className="animate-in fade-in duration-300 space-y-6">
      <h1 className="text-2xl font-bold tracking-tight">Portfolio</h1>

      {/* Summary Cards */}
      <div className="grid gap-4 sm:grid-cols-3">
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center gap-3">
              <div className="flex h-10 w-10 items-center justify-center rounded-full bg-secondary">
                <Coins className="h-5 w-5 text-primary" />
              </div>
              <div>
                <p className="text-sm text-muted-foreground">Invested</p>
                <p className="text-2xl font-bold tabular-nums">{totalInvested.toLocaleString()}</p>
              </div>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center gap-3">
              <div className="flex h-10 w-10 items-center justify-center rounded-full bg-secondary">
                <TrendingUp className="h-5 w-5 text-green-600" />
              </div>
              <div>
                <p className="text-sm text-muted-foreground">Potential Payout</p>
                <p className="text-2xl font-bold tabular-nums">{totalPotential.toLocaleString()}</p>
                <p className={`text-xs font-medium ${netProfit >= 0 ? "text-green-600" : "text-red-500"}`}>
                  {netProfit >= 0 ? "+" : ""}{netProfit.toLocaleString()} net
                </p>
              </div>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center gap-3">
              <div className="flex h-10 w-10 items-center justify-center rounded-full bg-secondary">
                <BarChart3 className="h-5 w-5 text-primary" />
              </div>
              <div>
                <p className="text-sm text-muted-foreground">Active Markets</p>
                <p className="text-2xl font-bold tabular-nums">{activeMarkets}</p>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Positions */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-base">
            <Briefcase className="h-4 w-4" />
            Active Positions
          </CardTitle>
        </CardHeader>
        <CardContent>
          {positions.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-12 text-center">
              <div className="mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-muted">
                <Briefcase className="h-8 w-8 text-muted-foreground" />
              </div>
              <p className="text-lg font-medium">No active positions</p>
              <p className="mt-1 text-sm text-muted-foreground">
                Place some bets and your positions will show up here.
              </p>
            </div>
          ) : (
            <div className="space-y-2">
              {positions.map((pos) => {
                const cost = Math.round(pos.shares);
                const profit = pos.potential_payout - cost;

                return (
                  <Link
                    key={`${pos.event_id}-${pos.outcome_id}`}
                    to={`/events/${pos.event_id}`}
                  >
                    <div className="flex items-center justify-between rounded-lg border px-3 py-3 hover:bg-muted/50 transition-colors">
                      <div className="min-w-0 flex-1">
                        <div className="font-medium truncate">
                          {pos.event_title}
                        </div>
                        <div className="mt-0.5 text-xs text-muted-foreground">
                          {pos.outcome_label} · {Math.round(pos.shares)} shares
                        </div>
                      </div>
                      <div className="ml-3 shrink-0 text-right">
                        <div className="text-sm font-semibold text-green-600 dark:text-green-400">
                          +{pos.potential_payout} pts
                        </div>
                        <div className="text-xs text-muted-foreground">
                          {profit >= 0 ? "+" : ""}{profit} net
                        </div>
                      </div>
                    </div>
                  </Link>
                );
              })}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
