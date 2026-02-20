import { useEffect, useState, useCallback, useRef } from "react";
import { getLeaderboard } from "@/lib/api";
import { useAuth } from "@/context/AuthContext";
import { useEventStream } from "@/hooks/useEventStream";
import type { LeaderboardEntry } from "@/lib/types";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { Trophy, Medal, Users, ChevronUp, ChevronDown } from "lucide-react";

const rankIcons: Record<number, { icon: typeof Trophy; color: string }> = {
  1: { icon: Trophy, color: "text-yellow-500" },
  2: { icon: Medal, color: "text-gray-400" },
  3: { icon: Medal, color: "text-amber-600" },
};

export default function LeaderboardPage() {
  const [entries, setEntries] = useState<LeaderboardEntry[]>([]);
  const [loading, setLoading] = useState(true);
  const [rankChanges, setRankChanges] = useState<Record<number, number>>({});
  const prevRanks = useRef<Record<number, number>>({});
  const { user } = useAuth();

  const fetchLeaderboard = useCallback(() => {
    getLeaderboard()
      .then((data) => {
        // Calculate rank changes from previous snapshot
        const prev = prevRanks.current;
        if (Object.keys(prev).length > 0) {
          const changes: Record<number, number> = {};
          for (const entry of data) {
            const oldRank = prev[entry.user_id];
            if (oldRank !== undefined && oldRank !== entry.rank) {
              changes[entry.user_id] = oldRank - entry.rank; // positive = moved up
            }
          }
          if (Object.keys(changes).length > 0) {
            setRankChanges(changes);
            // Clear movement indicators after 5 seconds
            setTimeout(() => setRankChanges({}), 5000);
          }
        }
        // Save current ranks for next comparison
        const current: Record<number, number> = {};
        for (const entry of data) {
          current[entry.user_id] = entry.rank;
        }
        prevRanks.current = current;
        setEntries(data);
      })
      .catch(() => {})
      .finally(() => setLoading(false));
  }, []);

  useEffect(() => {
    fetchLeaderboard();
  }, [fetchLeaderboard]);

  useEventStream(
    useCallback(
      (msg) => {
        if (
          msg.type === "odds_updated" ||
          msg.type === "event_resolved" ||
          msg.type === "connected"
        ) {
          fetchLeaderboard();
        }
      },
      [fetchLeaderboard]
    )
  );

  if (loading) {
    return (
      <div className="space-y-6">
        <Skeleton className="h-8 w-48" />
        <Card>
          <CardHeader>
            <Skeleton className="h-5 w-24" />
          </CardHeader>
          <CardContent className="space-y-2">
            {Array.from({ length: 6 }).map((_, i) => (
              <Skeleton key={i} className="h-11 w-full rounded-lg" />
            ))}
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="animate-in fade-in duration-300 space-y-6">
      <h1 className="text-2xl font-bold tracking-tight">Leaderboard</h1>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">Rankings</CardTitle>
        </CardHeader>
        <CardContent>
          {entries.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-12 text-center">
              <div className="mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-muted">
                <Users className="h-8 w-8 text-muted-foreground" />
              </div>
              <p className="text-lg font-medium">No players yet</p>
              <p className="mt-1 text-sm text-muted-foreground">
                Rankings will appear once players start trading.
              </p>
            </div>
          ) : (
            <div className="space-y-1">
              {entries.map((entry) => {
                const isMe = entry.user_id === user?.id;
                const rankInfo = rankIcons[entry.rank];
                const change = rankChanges[entry.user_id] ?? 0;

                return (
                  <div
                    key={entry.user_id}
                    className={`flex items-center justify-between rounded-lg px-3 py-2.5 transition-colors ${
                      isMe ? "bg-primary/5 ring-1 ring-primary/20" : "hover:bg-muted/50"
                    } ${change !== 0 ? "animate-in fade-in duration-500" : ""}`}
                  >
                    <div className="flex items-center gap-3">
                      <div className="flex h-8 w-8 items-center justify-center">
                        {rankInfo ? (
                          <rankInfo.icon
                            className={`h-5 w-5 ${rankInfo.color}`}
                          />
                        ) : (
                          <span className="text-sm font-medium text-muted-foreground">
                            #{entry.rank}
                          </span>
                        )}
                      </div>
                      <span className={`font-medium ${isMe ? "text-primary" : ""}`}>
                        {entry.username}
                        {isMe && (
                          <span className="ml-1.5 text-xs text-muted-foreground">
                            (you)
                          </span>
                        )}
                      </span>
                      {change > 0 && (
                        <span className="flex items-center gap-0.5 text-xs font-semibold text-green-600 animate-in slide-in-from-bottom-1 duration-300">
                          <ChevronUp className="h-3 w-3" />
                          {change}
                        </span>
                      )}
                      {change < 0 && (
                        <span className="flex items-center gap-0.5 text-xs font-semibold text-red-500 animate-in slide-in-from-top-1 duration-300">
                          <ChevronDown className="h-3 w-3" />
                          {Math.abs(change)}
                        </span>
                      )}
                    </div>
                    <span className="font-semibold tabular-nums">
                      {entry.balance.toLocaleString()} pts
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
