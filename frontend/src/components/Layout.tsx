import { Link, Outlet, useLocation, useNavigate } from "react-router-dom";
import { useAuth } from "@/context/AuthContext";
import { useEventStream } from "@/hooks/useEventStream";
import { useTheme } from "next-themes";
import { Button } from "@/components/ui/button";
import { toast } from "sonner";
import {
  TrendingUp,
  Trophy,
  History,
  Shield,
  LogOut,
  Coins,
  Menu,
  X,
  Sun,
  Moon,
  Grid3X3,
  Rss,
} from "lucide-react";
import { useState, useCallback, useEffect, useRef } from "react";

const navItems = [
  { to: "/events", label: "Markets", icon: TrendingUp },
  { to: "/leaderboard", label: "Leaderboard", icon: Trophy },
  { to: "/history", label: "History", icon: History },
  { to: "/activity", label: "Feed", icon: Rss },
];

export default function Layout() {
  const { user, logout, refreshUser } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();
  const { theme, setTheme } = useTheme();
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);
  const balanceRef = useRef<HTMLDivElement>(null);
  const prevBalance = useRef(user?.balance);

  // Flash balance pill on change
  useEffect(() => {
    if (
      prevBalance.current !== undefined &&
      user?.balance !== prevBalance.current &&
      balanceRef.current
    ) {
      balanceRef.current.classList.remove("balance-updated");
      void balanceRef.current.offsetWidth;
      balanceRef.current.classList.add("balance-updated");
    }
    prevBalance.current = user?.balance;
  }, [user?.balance]);

  const toggleTheme = () => {
    setTheme(theme === "dark" ? "light" : "dark");
  };

  useEventStream(
    useCallback(
      (msg) => {
        if (msg.type === "event_created") {
          const d = msg.data as { title: string };
          toast("New market created", { description: d.title });
        } else if (msg.type === "user_resolved") {
          const d = msg.data as {
            won: boolean;
            payout: number;
            refund: boolean;
            title: string;
          };
          if (d.refund) {
            toast("Event refunded", {
              description: `You got ${d.payout} pts back — ${d.title}`,
            });
          } else if (d.won) {
            toast.success(`You won +${d.payout} pts!`, {
              description: d.title,
            });
          } else {
            toast("Event resolved — you lost", {
              description: d.title,
            });
          }
          refreshUser();
        } else if (msg.type === "event_resolved") {
          const d = msg.data as { title: string; winner_label: string };
          toast("Market resolved", {
            description: `${d.title} — Winner: ${d.winner_label}`,
          });
          refreshUser();
        }
      },
      [refreshUser]
    )
  );

  const handleLogout = () => {
    logout();
    navigate("/login");
  };

  return (
    <div className="min-h-screen bg-background">
      <header className="sticky top-0 z-50 border-b bg-card/80 backdrop-blur-sm">
        <div className="mx-auto flex h-14 max-w-5xl items-center justify-between px-4">
          <div className="flex items-center gap-6">
            <Link
              to="/events"
              className="text-lg font-bold tracking-tight text-primary"
            >
              Paul-y Market
            </Link>

            <nav className="hidden items-center gap-1 md:flex">
              {navItems.map((item) => {
                const active = location.pathname.startsWith(item.to);
                return (
                  <Link key={item.to} to={item.to}>
                    <Button
                      variant={active ? "secondary" : "ghost"}
                      size="sm"
                      className="gap-1.5"
                    >
                      <item.icon className="h-4 w-4" />
                      {item.label}
                    </Button>
                  </Link>
                );
              })}
              {user?.bingo && (
                <Link to="/bingo">
                  <Button
                    variant={
                      location.pathname.startsWith("/bingo")
                        ? "secondary"
                        : "ghost"
                    }
                    size="sm"
                    className="gap-1.5"
                  >
                    <Grid3X3 className="h-4 w-4" />
                    Bingo
                  </Button>
                </Link>
              )}
              {user?.is_admin && (
                <Link to="/admin">
                  <Button
                    variant={
                      location.pathname.startsWith("/admin")
                        ? "secondary"
                        : "ghost"
                    }
                    size="sm"
                    className="gap-1.5"
                  >
                    <Shield className="h-4 w-4" />
                    Admin
                  </Button>
                </Link>
              )}
            </nav>
          </div>

          <div className="flex items-center gap-2">
            <div
              ref={balanceRef}
              className="flex items-center gap-1.5 rounded-full bg-secondary px-3 py-1.5 text-sm font-medium"
            >
              <Coins className="h-3.5 w-3.5 text-primary" />
              <span className="tabular-nums">
                {user?.balance.toLocaleString() ?? 0}
              </span>
            </div>

            <span className="hidden text-sm text-muted-foreground md:inline">
              {user?.username}
            </span>

            <Button
              variant="ghost"
              size="icon"
              className="h-8 w-8"
              onClick={toggleTheme}
            >
              <Sun className="h-4 w-4 rotate-0 scale-100 transition-transform dark:-rotate-90 dark:scale-0" />
              <Moon className="absolute h-4 w-4 rotate-90 scale-0 transition-transform dark:rotate-0 dark:scale-100" />
              <span className="sr-only">Toggle theme</span>
            </Button>

            <Button
              variant="ghost"
              size="sm"
              onClick={handleLogout}
              className="hidden gap-1.5 md:flex"
            >
              <LogOut className="h-4 w-4" />
            </Button>

            <Button
              variant="ghost"
              size="sm"
              className="md:hidden"
              onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
            >
              {mobileMenuOpen ? (
                <X className="h-5 w-5" />
              ) : (
                <Menu className="h-5 w-5" />
              )}
            </Button>
          </div>
        </div>

        {mobileMenuOpen && (
          <nav className="animate-in slide-in-from-top-2 duration-200 border-t px-4 py-2 md:hidden">
            <div className="flex flex-col gap-1">
              {navItems.map((item) => {
                const active = location.pathname.startsWith(item.to);
                return (
                  <Link
                    key={item.to}
                    to={item.to}
                    onClick={() => setMobileMenuOpen(false)}
                  >
                    <Button
                      variant={active ? "secondary" : "ghost"}
                      size="sm"
                      className="w-full justify-start gap-2"
                    >
                      <item.icon className="h-4 w-4" />
                      {item.label}
                    </Button>
                  </Link>
                );
              })}
              {user?.bingo && (
                <Link to="/bingo" onClick={() => setMobileMenuOpen(false)}>
                  <Button
                    variant={
                      location.pathname.startsWith("/bingo")
                        ? "secondary"
                        : "ghost"
                    }
                    size="sm"
                    className="w-full justify-start gap-2"
                  >
                    <Grid3X3 className="h-4 w-4" />
                    Bingo
                  </Button>
                </Link>
              )}
              {user?.is_admin && (
                <Link to="/admin" onClick={() => setMobileMenuOpen(false)}>
                  <Button
                    variant={
                      location.pathname.startsWith("/admin")
                        ? "secondary"
                        : "ghost"
                    }
                    size="sm"
                    className="w-full justify-start gap-2"
                  >
                    <Shield className="h-4 w-4" />
                    Admin
                  </Button>
                </Link>
              )}
              <Button
                variant="ghost"
                size="sm"
                onClick={handleLogout}
                className="w-full justify-start gap-2 text-destructive"
              >
                <LogOut className="h-4 w-4" />
                Log out
              </Button>
            </div>
          </nav>
        )}
      </header>

      <main className="mx-auto max-w-5xl px-4 py-6 pb-16">
        <Outlet />
      </main>
    </div>
  );
}
