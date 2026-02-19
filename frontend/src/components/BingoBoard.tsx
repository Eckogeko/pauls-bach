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
  selectedPosition?: number | null;
  onSquareClick?: (position: number) => void;
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
  const [expandedSquare, setExpandedSquare] = useState<number | null>(null);

  const byPos = new Map<number, BingoSquare>();
  for (const sq of squares) {
    byPos.set(sq.position, sq);
  }

  return (
    <>
      <div className="grid grid-cols-5 gap-1 sm:gap-1.5 w-full max-w-lg mx-auto">
        {Array.from({ length: 25 }, (_, i) => {
          const sq = byPos.get(i);
          const isWinning =
            props.mode === "view" && props.winningPositions?.has(i);

          if (props.mode === "build") {
            const hasEvent = sq?.bingo_event_id;
            const label = hasEvent ? getSquareLabel(sq!, bingoEvents) : "";
            const isSelected = props.selectedPosition === i;

            return (
              <div
                key={i}
                className={cn(
                  "aspect-square border rounded-md flex items-center justify-center p-0.5 sm:p-1 text-center transition-colors relative cursor-pointer",
                  isSelected
                    ? "border-primary bg-primary/15 border-2 ring-2 ring-primary/30"
                    : dragOver === i
                      ? "border-primary bg-primary/10 border-dashed border-2"
                      : hasEvent
                        ? "border-primary/30 bg-primary/5"
                        : "border-border bg-card border-dashed"
                )}
                onClick={() => props.onSquareClick?.(i)}
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
                    <span className="text-[9px] sm:text-[11px] leading-tight line-clamp-3 pr-2.5">
                      {label}
                    </span>
                    <button
                      type="button"
                      className="absolute top-0 right-0 p-0.5 rounded hover:bg-destructive/10 text-muted-foreground hover:text-destructive transition-colors"
                      onClick={(e) => {
                        e.stopPropagation();
                        props.onSquareRemove(i);
                      }}
                    >
                      <X className="h-2.5 w-2.5 sm:h-3 sm:w-3" />
                    </button>
                  </>
                ) : (
                  <span className="text-[9px] sm:text-[10px] text-muted-foreground/50">
                    {isSelected ? "Pick below" : "Tap"}
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
                "aspect-square border rounded-md flex items-center justify-center p-0.5 sm:p-1.5 text-center transition-all cursor-pointer select-none",
                sq?.resolved
                  ? isWinning
                    ? "bg-green-500/30 border-green-500 text-green-700 dark:text-green-300 font-semibold ring-1 ring-green-500/50 bingo-winning"
                    : "bg-green-500/15 border-green-500/50 text-green-700 dark:text-green-400"
                  : "border-border bg-card"
              )}
              onClick={() =>
                setExpandedSquare(expandedSquare === i ? null : i)
              }
            >
              <span className="text-[9px] sm:text-[11px] leading-tight line-clamp-3">
                {label}
              </span>
            </div>
          );
        })}
      </div>

      {/* Tap-to-expand overlay for view mode */}
      {props.mode === "view" && expandedSquare !== null && (() => {
        const sq = byPos.get(expandedSquare);
        const label = sq ? getSquareLabel(sq, bingoEvents) : "";
        if (!label) return null;
        return (
          <div
            className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-6"
            onClick={() => setExpandedSquare(null)}
          >
            <div
              className={cn(
                "rounded-xl px-6 py-5 max-w-sm w-full text-center shadow-lg animate-in zoom-in-95 fade-in duration-150",
                sq?.resolved
                  ? "bg-green-50 dark:bg-green-950 border-2 border-green-500"
                  : "bg-card border-2 border-border"
              )}
              onClick={(e) => e.stopPropagation()}
            >
              <p className={cn(
                "text-base font-medium leading-snug",
                sq?.resolved ? "text-green-700 dark:text-green-300" : ""
              )}>
                {label}
              </p>
              {sq?.resolved && (
                <p className="mt-2 text-sm text-green-600 dark:text-green-400 font-medium">Resolved</p>
              )}
            </div>
          </div>
        );
      })()}
    </>
  );
}
