import { useState } from "react";
import type { BingoSquare, BingoEvent } from "@/lib/types";
import { cn } from "@/lib/utils";
import { X } from "lucide-react";

interface BingoBoardViewProps {
  mode: "view";
  squares: BingoSquare[];
  bingoEvents: BingoEvent[];
  winningPositions?: Set<number>;
}

interface BingoBoardBuildProps {
  mode: "build";
  squares: BingoSquare[];
  bingoEvents: BingoEvent[];
  onSquareChange: (position: number, square: Partial<BingoSquare>) => void;
  onSquareRemove: (position: number) => void;
}

type BingoBoardProps = BingoBoardViewProps | BingoBoardBuildProps;

function getSquareLabel(
  sq: BingoSquare,
  bingoEvents: BingoEvent[]
): string {
  if (sq.bingo_event_id) {
    const ev = bingoEvents.find((e) => e.id === sq.bingo_event_id);
    return ev?.title ?? `Event #${sq.bingo_event_id}`;
  }
  return "";
}

export default function BingoBoard(props: BingoBoardProps) {
  const { squares, bingoEvents } = props;
  const [dragOver, setDragOver] = useState<number | null>(null);

  const byPos = new Map<number, BingoSquare>();
  for (const sq of squares) {
    byPos.set(sq.position, sq);
  }

  return (
    <div className="grid grid-cols-5 gap-1.5 w-full max-w-lg mx-auto">
      {Array.from({ length: 25 }, (_, i) => {
        const sq = byPos.get(i);
        const isWinning =
          props.mode === "view" && props.winningPositions?.has(i);

        if (props.mode === "build") {
          const hasEvent = sq?.bingo_event_id;
          const label = hasEvent ? getSquareLabel(sq!, bingoEvents) : "";

          return (
            <div
              key={i}
              className={cn(
                "aspect-square border rounded-md flex items-center justify-center p-1 text-center transition-colors relative",
                dragOver === i
                  ? "border-primary bg-primary/10 border-dashed border-2"
                  : hasEvent
                    ? "border-primary/30 bg-primary/5"
                    : "border-border bg-card border-dashed"
              )}
              onDragOver={(e) => {
                e.preventDefault();
                setDragOver(i);
              }}
              onDragEnter={(e) => {
                e.preventDefault();
                setDragOver(i);
              }}
              onDragLeave={() => setDragOver(null)}
              onDrop={(e) => {
                e.preventDefault();
                setDragOver(null);
                const eventId = Number(e.dataTransfer.getData("text/plain"));
                if (eventId) {
                  props.onSquareChange(i, { bingo_event_id: eventId });
                }
              }}
            >
              {hasEvent ? (
                <>
                  <span className="text-[11px] leading-tight line-clamp-3 pr-3">
                    {label}
                  </span>
                  <button
                    type="button"
                    className="absolute top-0.5 right-0.5 p-0.5 rounded hover:bg-destructive/10 text-muted-foreground hover:text-destructive transition-colors"
                    onClick={() => props.onSquareRemove(i)}
                  >
                    <X className="h-3 w-3" />
                  </button>
                </>
              ) : (
                <span className="text-[10px] text-muted-foreground/50">
                  Drop here
                </span>
              )}
            </div>
          );
        }

        // View mode
        const label = sq ? getSquareLabel(sq, bingoEvents) : "";
        return (
          <div
            key={i}
            className={cn(
              "aspect-square border rounded-md flex items-center justify-center p-1.5 text-center transition-colors",
              sq?.resolved
                ? isWinning
                  ? "bg-green-500/30 border-green-500 text-green-700 dark:text-green-300 font-semibold"
                  : "bg-green-500/15 border-green-500/50 text-green-700 dark:text-green-400"
                : "border-border bg-card"
            )}
          >
            <span className="text-[11px] leading-tight line-clamp-3">
              {label}
            </span>
          </div>
        );
      })}
    </div>
  );
}
