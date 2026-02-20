import { describe, it, expect } from "vitest";
import { parsePredictions } from "./parsePredictions";

// ── Helpers ──────────────────────────────────────────────────────────────

function makeForecastItem(date: string, close: number) {
  return { date, close };
}

// ── Null / undefined inputs ───────────────────────────────────────────────

describe("parsePredictions — empty / null inputs", () => {
  it("returns empty arrays for null input", () => {
    const result = parsePredictions(null as never);
    expect(result.forecast).toEqual([]);
    expect(result.history).toEqual([]);
  });

  it("returns empty arrays for undefined input", () => {
    const result = parsePredictions(undefined as never);
    expect(result.forecast).toEqual([]);
    expect(result.history).toEqual([]);
  });
});

// ── Array format ──────────────────────────────────────────────────────────

describe("parsePredictions — array format", () => {
  it("parses a plain array of prediction items", () => {
    const raw = [
      { date: "2025-01-06", close: 150 },
      { date: "2025-01-07", close: 151 },
    ];
    const { forecast, history } = parsePredictions(raw as never);
    expect(forecast).toHaveLength(2);
    expect(forecast[0]).toEqual({ date: "2025-01-06", close: 150 });
    expect(history).toHaveLength(0);
  });

  it("skips items missing date or close", () => {
    const raw = [
      { date: "2025-01-06", close: 150 },
      { close: 151 },             // no date → skip
      { date: "2025-01-08" },     // no close → skip
    ];
    const { forecast } = parsePredictions(raw as never);
    expect(forecast).toHaveLength(1);
  });

  it("accepts 'price' as alias for 'close'", () => {
    const raw = [{ date: "2025-01-06", price: 149.5 }];
    const { forecast } = parsePredictions(raw as never);
    expect(forecast[0].close).toBe(149.5);
  });

  it("accepts 'dt' as alias for 'date'", () => {
    const raw = [{ dt: "2025-01-06", close: 149.5 }];
    const { forecast } = parsePredictions(raw as never);
    expect(forecast[0].date).toBe("2025-01-06");
  });
});

// ── Dict format ───────────────────────────────────────────────────────────

describe("parsePredictions — dict format", () => {
  it("prefers 'full_forecast' key", () => {
    const raw = {
      full_forecast: [
        { date: "2025-01-06", close: 150 },
        { date: "2025-01-07", close: 151 },
        { date: "2025-01-08", close: 152 },
      ],
      week: [{ date: "2025-01-06", close: 150 }],
    };
    const { forecast } = parsePredictions(raw as never);
    expect(forecast).toHaveLength(3);
  });

  it("falls back to 'forecast' key when full_forecast absent", () => {
    const raw = {
      forecast: [{ date: "2025-01-06", close: 150 }],
    };
    const { forecast } = parsePredictions(raw as never);
    expect(forecast).toHaveLength(1);
  });

  it("falls back to 'predictions' key", () => {
    const raw = {
      predictions: [{ date: "2025-01-06", close: 150 }],
    };
    const { forecast } = parsePredictions(raw as never);
    expect(forecast).toHaveLength(1);
  });

  it("falls back to first array found in the dict", () => {
    const raw = {
      some_other_key: [{ date: "2025-01-06", close: 150 }],
    };
    const { forecast } = parsePredictions(raw as never);
    expect(forecast).toHaveLength(1);
  });

  it("extracts history from dict.history", () => {
    const raw = {
      full_forecast: [{ date: "2025-01-10", close: 155 }],
      history: [
        { date: "2024-12-30", close: 148 },
        { date: "2024-12-31", close: 149 },
      ],
    };
    const { forecast, history } = parsePredictions(raw as never);
    expect(forecast).toHaveLength(1);
    expect(history).toHaveLength(2);
    expect(history[0]).toEqual({ date: "2024-12-30", close: 148 });
  });

  it("returns empty forecast when no array found", () => {
    const raw = { some_string: "value", some_number: 42 };
    const { forecast, history } = parsePredictions(raw as never);
    expect(forecast).toHaveLength(0);
    expect(history).toHaveLength(0);
  });
});

// ── Multi-horizon dict (new structure) ───────────────────────────────────

describe("parsePredictions — multi-horizon dict (week/month/quarter)", () => {
  it("reads full_forecast (63-day quarter) when present", () => {
    const quarter = Array.from({ length: 63 }, (_, i) => ({
      date: `2025-01-${String(i + 6).padStart(2, "0")}`,
      close: 150 + i,
    }));
    const raw = {
      full_forecast: quarter,
      week: quarter.slice(0, 5),
      month: quarter.slice(0, 21),
      quarter,
    };
    const { forecast } = parsePredictions(raw as never);
    expect(forecast).toHaveLength(63);
  });
});

// ── String fallback ───────────────────────────────────────────────────────

describe("parsePredictions — string fallback", () => {
  it("parses date: $price lines", () => {
    const raw =
      "5-Day Price Forecast:\n  2025-01-06: $150.50\n  2025-01-07: $151.00\n";
    const { forecast } = parsePredictions(raw as never);
    expect(forecast).toHaveLength(2);
    expect(forecast[0].close).toBeCloseTo(150.5);
    expect(forecast[0].date).toBe("2025-01-06");
  });

  it("returns empty arrays for a plain string with no price lines", () => {
    const { forecast, history } = parsePredictions("no data" as never);
    expect(forecast).toHaveLength(0);
    expect(history).toHaveLength(0);
  });
});

// ── Type coercion ─────────────────────────────────────────────────────────

describe("parsePredictions — type coercion", () => {
  it("coerces numeric date to string", () => {
    const raw = [{ date: 20250106, close: 150 }];
    const { forecast } = parsePredictions(raw as never);
    expect(typeof forecast[0].date).toBe("string");
  });

  it("coerces string close to number", () => {
    const raw = [{ date: "2025-01-06", close: "150.5" }];
    const { forecast } = parsePredictions(raw as never);
    expect(typeof forecast[0].close).toBe("number");
    expect(forecast[0].close).toBeCloseTo(150.5);
  });
});
