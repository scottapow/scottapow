export interface ITrack {
  id: string;
  slug: string;
  name: string;
  description: string;
  things: Array<ITrackThing>;
}

export type Type = 'number' | 'text';
type GoalPrimitive = number;
export type Goal<P = GoalPrimitive> = P;// | { su: P, m: P, tu: P, w: P, th: P, f: P, sa: P };
export interface ITrackThing {
  id: number;
  name: string;
  trackId: ITrack['id'];
  description?: string;
  type: Type;
  goal: Goal;
  entries: Array<ITrackEntry>;
}

export interface ITrackEntry {
  id: string;
  thingId: ITrackThing['id'],
  /** GMT/UTC string (e.g. `"Fri, 10 Oct 2025 20:31:44 GMT"`) */
  startAt: Date;
  /** GMT/UTC string (e.g. `"Fri, 10 Oct 2025 20:31:44 GMT"`) */
  endAt: Date | null;
  /** GMT/UTC string (e.g. `"Fri, 10 Oct 2025 20:31:44 GMT"`) */
  loggedAt: string;
  value: number | string;
  note: string | null;
  // relatesTo?: ITrackEntry['id'];
  // relationType?: unknown;
}