export interface ITrack {
  id: string;
  slug: string;
  name: string;
  description: string;
  things: Array<ITrackThing>;
}

export type Type = 'number' | 'text';
type GoalPrimitive = number | null;
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
  startTime: string;
  endTime: string | null;
  value: number | string;
  notes: string | null;
  // relatesTo?: ITrackEntry['id'];
  // relationType?: unknown;
}