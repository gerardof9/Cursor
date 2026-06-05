export type GuardPeriod = {
  id: string;
  person: string;
  start: Date;
  end: Date;
  sourceRow?: number;
};

export type ScheduleData = {
  periods: GuardPeriod[];
  people: string[];
  skippedRowCount: number;
  warnings: string[];
  loadedAt: string;
};

export type ScheduleDataSerialized = {
  periods: Array<{
    id: string;
    person: string;
    start: string;
    end: string;
    sourceRow?: number;
  }>;
  people: string[];
  skippedRowCount: number;
  warnings: string[];
  loadedAt: string;
};

export type ParseRowErrorReason =
  | "missing_person"
  | "invalid_dates"
  | "incomplete_week"
  | "bad_headers";
