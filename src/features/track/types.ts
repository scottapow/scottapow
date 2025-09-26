export interface ITrack {
  id: string;
  name: string;
  description: string;
  things: Array<ITrackThing>;
}

export interface ITrackThing {
  id: string;
  trackId: ITrack['id'];
  name: string;
  description: string;
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