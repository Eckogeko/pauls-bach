import { useState, useEffect, useCallback } from "react";
import type { BingoEvent, BingoBoard as BingoBoardType, BingoSquare, BingoWinner } from "@/lib/types";
import { getBingoEvents, getBingoBoard, createBingoBoard, getBingoWinners, getAllBingoBoards } from "@/lib/api";
import { useEventStream } from "@/hooks/useEventStream";
import BingoBoard from "@/components/BingoBoard";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { toast } from "sonner";
import { Loader2, Trophy, GripVertical } from "lucide-react";

const CUSTOM_POSITIONS = new Set([0, 4, 12, 20, 24]);

function getWinningPositions(winners: BingoWinner[]): Set<number> {
  const lines: Record<string, number[]> = {
    "row-0": [0, 1, 2, 3, 4],
    "row-1": [5, 6, 7, 8, 9],
    "row-2": [10, 11, 12, 13, 14],
    "row-3": [15, 16, 17, 18, 19],
    "row-4": [20, 21, 22, 23, 24],
    "col-0": [0, 5, 10, 15, 20],
    "col-1": [1, 6, 11, 16, 21],
    "col-2": [2, 7, 12, 17, 22],
    "col-3": [3, 8, 13, 18, 23],
    "col-4": [4, 9, 14, 19, 24],
    "diag-0": [0, 6, 12, 18, 24],
    "diag-1": [4, 8, 12, 16, 20],
  };
  const positions = new Set<number>();
  for (const w of winners) {
    const linePositions = lines[w.line];
    if (linePositions) {
      for (const p of linePositions) positions.add(p);
    }
  }
  return positions;
}

export default function BingoPage() {
  const [bingoEvents, setBingoEvents] = useState<BingoEvent[]>([]);
  const [board, setBoard] = useState<BingoBoardType | null>(null);
  const [boardWinners, setBoardWinners] = useState<BingoWinner[]>([]);
  const [allWinners, setAllWinners] = useState<BingoWinner[]>([]);
  const [allBoards, setAllBoards] = useState<(BingoBoardType & { username: string; winners: BingoWinner[] })[]>([]);
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [hasBoard, setHasBoard] = useState(false);

  const [buildSquares, setBuildSquares] = useState<Map<number, Partial<BingoSquare>>>(() => new Map());

  const fetchData = useCallback(async () => {
    try {
      const [events, winners] = await Promise.all([
        getBingoEvents(),
        getBingoWinners(),
      ]);
      setBingoEvents(events);
      setAllWinners(winners);

      try {
        const res = await getBingoBoard();
        setBoard(res.board);
        setBoardWinners(res.winners ?? []);
        setHasBoard(true);

        // Fetch all boards once user has their own
        try {
          const boards = await getAllBingoBoards();
          setAllBoards(boards);
        } catch {
          // ignore
        }
      } catch {
        setHasBoard(false);
      }
    } catch {
      // ignore
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  useEventStream(
    useCallback(
      (msg) => {
        if (msg.type === "bingo_resolved" || msg.type === "bingo_winner") {
          fetchData();
        }
        if (msg.type === "bingo_winner") {
          const d = msg.data as { username: string; message: string };
          toast.success(d.message || `${d.username} got BINGO!`);
        }
      },
      [fetchData]
    )
  );

  const handleSquareChange = (position: number, update: Partial<BingoSquare>) => {
    setBuildSquares((prev) => {
      const next = new Map(prev);
      const existing = next.get(position) ?? { position, resolved: false };
      next.set(position, { ...existing, ...update, position });
      return next;
    });
  };

  const handleSquareRemove = (position: number) => {
    setBuildSquares((prev) => {
      const next = new Map(prev);
      next.delete(position);
      return next;
    });
  };

  // Compute which event IDs are already placed on the board
  const usedEventIds = new Set<number>();
  for (const [, sq] of buildSquares) {
    if (sq.bingo_event_id) usedEventIds.add(sq.bingo_event_id);
  }

  const availableEvents = bingoEvents.filter((ev) => !usedEventIds.has(ev.id));

  const handleSubmit = async () => {
    const squares: BingoSquare[] = [];
    for (let i = 0; i < 25; i++) {
      const data = buildSquares.get(i);
      if (CUSTOM_POSITIONS.has(i)) {
        if (!data?.custom_text?.trim()) {
          toast.error(`Custom square at position ${i} needs text`);
          return;
        }
        squares.push({
          position: i,
          custom_text: data.custom_text.trim(),
          resolved: false,
        });
      } else {
        if (!data?.bingo_event_id) {
          toast.error("All event squares must have an event selected");
          return;
        }
        squares.push({
          position: i,
          bingo_event_id: data.bingo_event_id,
          resolved: false,
        });
      }
    }

    setSubmitting(true);
    try {
      await createBingoBoard(squares);
      toast.success("Bingo board created!");
      fetchData();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to create board");
    } finally {
      setSubmitting(false);
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center py-20">
        <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
      </div>
    );
  }

  const builderSquares: BingoSquare[] = Array.from({ length: 25 }, (_, i) => {
    const data = buildSquares.get(i);
    return {
      position: i,
      bingo_event_id: data?.bingo_event_id,
      custom_text: data?.custom_text,
      resolved: false,
    };
  });

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold tracking-tight">Bingo</h1>

      {hasBoard && board ? (
        <>
          {boardWinners.length > 0 && (
            <Card className="border-green-500/50 bg-green-500/5">
              <CardContent className="flex items-center gap-2 py-3">
                <Trophy className="h-5 w-5 text-green-600" />
                <span className="font-semibold text-green-700 dark:text-green-400">
                  BINGO! You completed: {boardWinners.map((w) => w.line).join(", ")}
                </span>
              </CardContent>
            </Card>
          )}

          <Card>
            <CardHeader>
              <CardTitle className="text-base">Your Board</CardTitle>
              <CardDescription>
                Resolved squares are highlighted. Corners and center are custom.
              </CardDescription>
            </CardHeader>
            <CardContent>
              <BingoBoard
                mode="view"
                squares={board.squares}
                bingoEvents={bingoEvents}
                winningPositions={getWinningPositions(boardWinners)}
              />
            </CardContent>
          </Card>
        </>
      ) : (
        <Card>
          <CardHeader>
            <CardTitle className="text-base">Build Your Board</CardTitle>
            <CardDescription>
              Drag events from the list into the grid. Corners and center (highlighted) are custom text.
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            {bingoEvents.length === 0 ? (
              <p className="text-sm text-muted-foreground py-4 text-center">
                No bingo events available yet. Ask the admin to create some!
              </p>
            ) : (
              <>
                <div className="flex gap-4 flex-col md:flex-row">
                  {/* Sidebar: draggable event list */}
                  <div className="md:w-48 shrink-0 space-y-1.5">
                    <div className="text-xs font-medium text-muted-foreground mb-2">
                      Events ({availableEvents.length} remaining)
                    </div>
                    <div className="max-h-[420px] overflow-y-auto space-y-1 pr-1">
                      {availableEvents.map((ev) => (
                        <div
                          key={ev.id}
                          draggable
                          onDragStart={(e) => {
                            e.dataTransfer.setData("text/plain", String(ev.id));
                            e.dataTransfer.effectAllowed = "move";
                          }}
                          className="flex items-start gap-1.5 rounded-md border px-2 py-1.5 text-xs cursor-grab active:cursor-grabbing hover:bg-accent transition-colors select-none"
                        >
                          <GripVertical className="h-3 w-3 text-muted-foreground shrink-0 mt-0.5" />
                          <span>{ev.title}</span>
                        </div>
                      ))}
                      {availableEvents.length === 0 && (
                        <p className="text-[11px] text-muted-foreground/60 text-center py-2">
                          All events placed!
                        </p>
                      )}
                    </div>
                  </div>

                  {/* Board grid */}
                  <div className="flex-1">
                    <BingoBoard
                      mode="build"
                      squares={builderSquares}
                      bingoEvents={bingoEvents}
                      onSquareChange={handleSquareChange}
                      onSquareRemove={handleSquareRemove}
                    />
                  </div>
                </div>

                <Button
                  className="w-full"
                  onClick={handleSubmit}
                  disabled={submitting}
                >
                  {submitting ? (
                    <Loader2 className="h-4 w-4 animate-spin" />
                  ) : (
                    "Lock In Board"
                  )}
                </Button>
              </>
            )}
          </CardContent>
        </Card>
      )}

      {hasBoard && allBoards.filter((b) => b.user_id !== board?.user_id).length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle className="text-base">Other Boards</CardTitle>
            <CardDescription>See how other players set up their boards.</CardDescription>
          </CardHeader>
          <CardContent className="space-y-6">
            {allBoards
              .filter((b) => b.user_id !== board?.user_id)
              .map((b) => (
                <div key={b.id} className="space-y-2">
                  <div className="flex items-center gap-2">
                    <span className="font-medium">{b.username}</span>
                    {b.winners && b.winners.length > 0 && (
                      <Badge variant="secondary" className="gap-1">
                        <Trophy className="h-3 w-3" /> BINGO
                      </Badge>
                    )}
                  </div>
                  <BingoBoard
                    mode="view"
                    squares={b.squares}
                    bingoEvents={bingoEvents}
                    winningPositions={getWinningPositions(b.winners ?? [])}
                  />
                </div>
              ))}
          </CardContent>
        </Card>
      )}

      {allWinners.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle className="text-base">Bingo Winners</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-2">
              {allWinners.map((w) => (
                <div
                  key={w.id}
                  className="flex items-center justify-between rounded-lg border px-3 py-2"
                >
                  <div>
                    <span className="font-medium">{w.username}</span>
                    <span className="ml-2 text-xs text-muted-foreground">
                      {w.line}
                    </span>
                  </div>
                  <Badge variant="secondary">BINGO</Badge>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
}
