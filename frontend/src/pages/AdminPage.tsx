import { useState, useEffect } from "react";
import type { Event, BingoEvent } from "@/lib/types";
import {
  getEvents,
  createEvent,
  resolveEvent,
  getBingoEvents,
  createBingoEvent,
  resolveBingoEvent,
  getAdminUsers,
  setUserBingo,
} from "@/lib/api";
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
import { toast } from "sonner";
import { Switch } from "@/components/ui/switch";
import { Plus, X, CheckCircle2, Loader2, Grid3X3, Users } from "lucide-react";

export default function AdminPage() {
  const [events, setEvents] = useState<Event[]>([]);
  const [loading, setLoading] = useState(true);

  // Create event form
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [eventType, setEventType] = useState("binary");
  const [outcomes, setOutcomes] = useState(["Yes", "No"]);
  const [createLoading, setCreateLoading] = useState(false);
  const [createOpen, setCreateOpen] = useState(false);

  // Resolve event
  const [resolveOpen, setResolveOpen] = useState(false);
  const [resolveTarget, setResolveTarget] = useState<Event | null>(null);
  const [winnerOutcomeId, setWinnerOutcomeId] = useState<string>("");
  const [resolveLoading, setResolveLoading] = useState(false);

  // Users
  const [users, setUsers] = useState<{ id: number; username: string; is_admin: boolean; bingo: boolean }[]>([]);

  // Bingo
  const [bingoEvents, setBingoEvents] = useState<BingoEvent[]>([]);
  const [bingoTitle, setBingoTitle] = useState("");
  const [bingoCreateOpen, setBingoCreateOpen] = useState(false);
  const [bingoCreateLoading, setBingoCreateLoading] = useState(false);

  const fetchEvents = () => {
    getEvents()
      .then(setEvents)
      .catch(() => {})
      .finally(() => setLoading(false));
  };

  const fetchBingoEvents = () => {
    getBingoEvents().then(setBingoEvents).catch(() => {});
  };

  const fetchUsers = () => {
    getAdminUsers().then(setUsers).catch(() => {});
  };

  useEffect(() => {
    fetchEvents();
    fetchBingoEvents();
    fetchUsers();
  }, []);

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
      await createEvent(title.trim(), description.trim(), eventType, outcomes);
      toast.success("Event created");
      setTitle("");
      setDescription("");
      setEventType("binary");
      setOutcomes(["Yes", "No"]);
      setCreateOpen(false);
      fetchEvents();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to create");
    } finally {
      setCreateLoading(false);
    }
  };

  const handleResolve = async () => {
    if (!resolveTarget || !winnerOutcomeId) return;
    setResolveLoading(true);
    try {
      await resolveEvent(resolveTarget.id, Number(winnerOutcomeId));
      toast.success("Event resolved");
      setResolveOpen(false);
      setResolveTarget(null);
      setWinnerOutcomeId("");
      fetchEvents();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to resolve");
    } finally {
      setResolveLoading(false);
    }
  };

  const handleEventTypeChange = (type: string) => {
    setEventType(type);
    if (type === "binary") {
      setOutcomes(["Yes", "No"]);
    } else {
      setOutcomes(["", "", ""]);
    }
  };

  const openEvents = events.filter((e) => e.status === "open");
  const otherEvents = events.filter((e) => e.status !== "open");

  if (loading) {
    return (
      <div className="flex items-center justify-center py-20">
        <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold tracking-tight">Admin Panel</h1>

        <Dialog open={createOpen} onOpenChange={setCreateOpen}>
          <DialogTrigger asChild>
            <Button className="gap-1.5">
              <Plus className="h-4 w-4" />
              New Event
            </Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Create Event</DialogTitle>
              <DialogDescription>
                Create a new prediction market for players to trade on.
              </DialogDescription>
            </DialogHeader>
            <div className="space-y-4">
              <div className="space-y-2">
                <Label>Title</Label>
                <Input
                  placeholder="Will it rain tomorrow?"
                  value={title}
                  onChange={(e) => setTitle(e.target.value)}
                />
              </div>
              <div className="space-y-2">
                <Label>Description (optional)</Label>
                <Input
                  placeholder="Additional context..."
                  value={description}
                  onChange={(e) => setDescription(e.target.value)}
                />
              </div>
              <div className="space-y-2">
                <Label>Type</Label>
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
              <div className="space-y-2">
                <Label>Outcomes</Label>
                <div className="space-y-2">
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
                        disabled={eventType === "binary"}
                      />
                      {eventType === "multi" && outcomes.length > 2 && (
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
                  {eventType === "multi" && (
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => setOutcomes([...outcomes, ""])}
                    >
                      <Plus className="mr-1 h-3.5 w-3.5" />
                      Add Outcome
                    </Button>
                  )}
                </div>
              </div>
              <Button
                className="w-full"
                onClick={handleCreate}
                disabled={createLoading}
              >
                {createLoading ? (
                  <Loader2 className="h-4 w-4 animate-spin" />
                ) : (
                  "Create Event"
                )}
              </Button>
            </div>
          </DialogContent>
        </Dialog>
      </div>

      {/* Open events — can be resolved */}
      <Card>
        <CardHeader>
          <CardTitle className="text-base">Open Events</CardTitle>
          <CardDescription>Click resolve to settle an event</CardDescription>
        </CardHeader>
        <CardContent>
          {openEvents.length === 0 ? (
            <div className="py-6 text-center text-muted-foreground">
              No open events
            </div>
          ) : (
            <div className="space-y-2">
              {openEvents.map((event) => (
                <div
                  key={event.id}
                  className="flex items-center justify-between rounded-lg border px-3 py-2.5"
                >
                  <div>
                    <div className="font-medium">{event.title}</div>
                    <div className="text-xs text-muted-foreground">
                      {event.odds.map((o) => o.label).join(" / ")}
                    </div>
                  </div>
                  <Button
                    variant="outline"
                    size="sm"
                    className="gap-1"
                    onClick={() => {
                      setResolveTarget(event);
                      setWinnerOutcomeId("");
                      setResolveOpen(true);
                    }}
                  >
                    <CheckCircle2 className="h-3.5 w-3.5" />
                    Resolve
                  </Button>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      {/* Resolved/closed events */}
      {otherEvents.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle className="text-base">Past Events</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-2">
              {otherEvents.map((event) => (
                <div
                  key={event.id}
                  className="flex items-center justify-between rounded-lg border px-3 py-2.5"
                >
                  <div>
                    <div className="font-medium">{event.title}</div>
                    <div className="text-xs text-muted-foreground">
                      {event.odds.map((o) => o.label).join(" / ")}
                    </div>
                  </div>
                  <Badge variant="secondary">{event.status}</Badge>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      )}

      {/* User Permissions */}
      <Card>
        <CardHeader>
          <div>
            <CardTitle className="text-base flex items-center gap-1.5">
              <Users className="h-4 w-4" />
              User Permissions
            </CardTitle>
            <CardDescription>Toggle bingo access for users</CardDescription>
          </div>
        </CardHeader>
        <CardContent>
          {users.length === 0 ? (
            <div className="py-6 text-center text-muted-foreground">
              No users
            </div>
          ) : (
            <div className="space-y-2">
              {users.map((u) => (
                <div
                  key={u.id}
                  className="flex items-center justify-between rounded-lg border px-3 py-2.5"
                >
                  <div className="font-medium">
                    {u.username}
                    {u.is_admin && (
                      <Badge variant="secondary" className="ml-2">admin</Badge>
                    )}
                  </div>
                  <div className="flex items-center gap-2">
                    <span className="text-xs text-muted-foreground">Bingo</span>
                    <Switch
                      checked={u.bingo}
                      onCheckedChange={async (checked) => {
                        try {
                          await setUserBingo(u.id, checked);
                          setUsers((prev) =>
                            prev.map((x) =>
                              x.id === u.id ? { ...x, bingo: checked } : x
                            )
                          );
                        } catch (err) {
                          toast.error(
                            err instanceof Error ? err.message : "Failed to update"
                          );
                        }
                      }}
                    />
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      {/* Bingo Events */}
      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <div>
            <CardTitle className="text-base">Bingo Events</CardTitle>
            <CardDescription>Create and resolve bingo events</CardDescription>
          </div>
          <Dialog open={bingoCreateOpen} onOpenChange={setBingoCreateOpen}>
            <DialogTrigger asChild>
              <Button size="sm" variant="outline" className="gap-1.5">
                <Grid3X3 className="h-3.5 w-3.5" />
                New Bingo Event
              </Button>
            </DialogTrigger>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>Create Bingo Event</DialogTitle>
                <DialogDescription>
                  Add an event to the bingo pool for players to put on their boards.
                </DialogDescription>
              </DialogHeader>
              <div className="space-y-4">
                <div className="space-y-2">
                  <Label>Title</Label>
                  <Input
                    placeholder="Paul wears a hat"
                    value={bingoTitle}
                    onChange={(e) => setBingoTitle(e.target.value)}
                  />
                </div>
                <Button
                  className="w-full"
                  disabled={bingoCreateLoading}
                  onClick={async () => {
                    if (!bingoTitle.trim()) {
                      toast.error("Title is required");
                      return;
                    }
                    setBingoCreateLoading(true);
                    try {
                      await createBingoEvent(bingoTitle.trim());
                      toast.success("Bingo event created");
                      setBingoTitle("");
                      setBingoCreateOpen(false);
                      fetchBingoEvents();
                    } catch (err) {
                      toast.error(
                        err instanceof Error ? err.message : "Failed to create"
                      );
                    } finally {
                      setBingoCreateLoading(false);
                    }
                  }}
                >
                  {bingoCreateLoading ? (
                    <Loader2 className="h-4 w-4 animate-spin" />
                  ) : (
                    "Create"
                  )}
                </Button>
              </div>
            </DialogContent>
          </Dialog>
        </CardHeader>
        <CardContent>
          {bingoEvents.length === 0 ? (
            <div className="py-6 text-center text-muted-foreground">
              No bingo events yet
            </div>
          ) : (
            <div className="space-y-2">
              {bingoEvents.map((be) => (
                <div
                  key={be.id}
                  className="flex items-center justify-between rounded-lg border px-3 py-2.5"
                >
                  <div className="font-medium">{be.title}</div>
                  {be.resolved ? (
                    <Badge variant="secondary">Resolved</Badge>
                  ) : (
                    <Button
                      variant="outline"
                      size="sm"
                      className="gap-1"
                      onClick={async () => {
                        try {
                          await resolveBingoEvent(be.id);
                          toast.success("Bingo event resolved");
                          fetchBingoEvents();
                        } catch (err) {
                          toast.error(
                            err instanceof Error
                              ? err.message
                              : "Failed to resolve"
                          );
                        }
                      }}
                    >
                      <CheckCircle2 className="h-3.5 w-3.5" />
                      Resolve
                    </Button>
                  )}
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      {/* Resolve Dialog */}
      <Dialog open={resolveOpen} onOpenChange={setResolveOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Resolve Event</DialogTitle>
            <DialogDescription>
              Select the winning outcome for "{resolveTarget?.title}"
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <Select value={winnerOutcomeId} onValueChange={setWinnerOutcomeId}>
              <SelectTrigger>
                <SelectValue placeholder="Select winner..." />
              </SelectTrigger>
              <SelectContent>
                {resolveTarget?.odds.map((o) => (
                  <SelectItem key={o.outcome_id} value={String(o.outcome_id)}>
                    {o.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <Button
              className="w-full"
              onClick={handleResolve}
              disabled={!winnerOutcomeId || resolveLoading}
            >
              {resolveLoading ? (
                <Loader2 className="h-4 w-4 animate-spin" />
              ) : (
                "Confirm Resolution"
              )}
            </Button>
          </div>
        </DialogContent>
      </Dialog>
    </div>
  );
}
