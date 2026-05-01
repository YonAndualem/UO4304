"use client";

import { useEffect, useState } from "react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";

export interface ActionOption {
  value: string;
  label: string;
  requiresNotes: boolean;
  variant: "default" | "destructive" | "outline";
}

interface Props {
  open: boolean;
  onClose: () => void;
  onSubmit: (action: string, notes: string) => Promise<void>;
  actions: ActionOption[];
  title: string;
  preSelected?: string;
}

export function ActionModal({ open, onClose, onSubmit, actions, title, preSelected }: Props) {
  const preOption = preSelected ? (actions.find((a) => a.value === preSelected) ?? null) : null;
  const [selected, setSelected] = useState<ActionOption | null>(preOption);
  const [notes, setNotes] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  // Sync preSelected when it changes (e.g. different row button clicked)
  useEffect(() => {
    setSelected(preOption);
    setNotes("");
    setError("");
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [preSelected, open]);

  const requiresNotes = selected?.requiresNotes ?? false;

  async function handleSubmit() {
    if (!selected) return;
    if (requiresNotes && notes.trim() === "") {
      setError("Notes are required for this action.");
      return;
    }
    setError("");
    setLoading(true);
    try {
      await onSubmit(selected.value, notes.trim());
      handleClose();
    } catch (e) {
      setError(e instanceof Error ? e.message : "An error occurred.");
    } finally {
      setLoading(false);
    }
  }

  function handleClose() {
    setSelected(null);
    setNotes("");
    setError("");
    onClose();
  }

  return (
    <Dialog open={open} onOpenChange={(o: boolean) => !o && handleClose()}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
        </DialogHeader>

        <div className="space-y-4 py-2">
          {!preOption && (
            <div className="flex flex-wrap gap-2">
              {actions.map((a) => (
                <Button
                  key={a.value}
                  variant={selected?.value === a.value ? "default" : "outline"}
                  onClick={() => { setSelected(a); setError(""); }}
                  className="flex-1"
                >
                  {a.label}
                </Button>
              ))}
            </div>
          )}

          {(selected || preOption) && (
            <div className="space-y-1.5">
              <Label htmlFor="action-notes">
                Notes {requiresNotes && <span className="text-red-500">*</span>}
              </Label>
              <Textarea
                id="action-notes"
                placeholder="Add notes…"
                value={notes}
                onChange={(e) => setNotes(e.target.value.slice(0, 1000))}
                rows={3}
              />
            </div>
          )}

          {error && <p className="text-sm text-red-600">{error}</p>}
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={handleClose} disabled={loading}>
            Cancel
          </Button>
          <Button onClick={handleSubmit} disabled={!selected || loading}>
            {loading ? "Submitting…" : "Confirm"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
