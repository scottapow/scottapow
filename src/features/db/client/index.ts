import type { ITrack, ITrackEntry, ITrackThing } from "../../track/types";
import type { IStore } from "./types";

const V = 2; // cannot be a float
export const OS_NAME_TRACKS = 'tracks' as const;
export const OS_NAME_THINGS = 'things' as const;
export const OS_NAME_ENTRIES = 'entries' as const;
export const thingCreateEventType = `${OS_NAME_THINGS}Add` as const;
export const entryCreateEventType = `${OS_NAME_ENTRIES}Add` as const;
export const entryUpdateEventType = `${OS_NAME_ENTRIES}Update` as const;
export const entryDeleteEventType = `${OS_NAME_ENTRIES}Delete` as const;

const DEBUG = false;

export class ClientDB {
  DB?: IDBDatabase;
  state: 'ready' | 'error' | 'idle' = 'idle';

  private async add<T extends any>(osName: string, data: T): Promise<IDBValidKey | Error> {
    let tx = this.DB?.transaction(osName, "readwrite");
    return new Promise((resolve, reject) => {

      if (!tx) {
        reject(new Error('Failed to create transaction'));
        return;
      }

      let os = tx?.objectStore(osName);
      let rq = os?.add(data);

      tx.oncomplete = (ev) => {
        resolve(rq.result);
      }

      tx.onerror = (ev) => {
        let event = ev as Event & { target: IDBRequest };
        reject(new Error('Failed to add thing', { cause: event.target.error }));
      };
    });
  }

  async connect(name: string): Promise<typeof this> {
    return new Promise((resolve, reject) => {
      const request = window.indexedDB.open(name, V);

      request.onerror = () => {
        if (DEBUG) console.log('onerror');
        if (request.error) {
          if (DEBUG) console.log('request error');
          reject(request.error.message);
        } else {
          reject(new Error('Unknown error on IndexedDB request'));
        }
      }

      request.onsuccess = (ev) => {
        if (DEBUG) console.log('onsuccess');
        this.DB = (ev.target as EventTarget & { result: IDBDatabase })?.result;
        resolve(this);
      }

      request.onupgradeneeded = (ev) => {
        if (DEBUG) console.log('onupgradeneeded');
        this.DB = (ev.target as EventTarget & { result: IDBDatabase })?.result;
        try {
          const tracksOS: IDBObjectStore = this.DB?.createObjectStore(OS_NAME_TRACKS, { keyPath: 'id' })!;
          tracksOS?.createIndex('slug', 'slug', { unique: true }); // PK - used for routing
          tracksOS?.createIndex('id', 'id', { unique: true }); // PK

          const thingsOS: IDBObjectStore = this.DB?.createObjectStore(OS_NAME_THINGS, { keyPath: 'id', autoIncrement: true })!;
          thingsOS?.createIndex('id', 'id', { unique: true }); // PK
          thingsOS?.createIndex('trackId', 'trackId', { unique: false }); // FK

          const entriesOD: IDBObjectStore = this.DB?.createObjectStore(OS_NAME_ENTRIES, { keyPath: 'id', autoIncrement: true })!;
          entriesOD?.createIndex('id', 'id', { unique: true }); // PK
          entriesOD?.createIndex('thingId', 'thingId', { unique: false }); // FK
          entriesOD?.createIndex('startAt', 'startAt', { unique: false }); // FK
          entriesOD?.createIndex('endAt', 'endAt', { unique: false }); // FK
          entriesOD?.createIndex('note', 'note', { unique: false }); // FK

          tracksOS.transaction.onerror = (ev) => {
            reject(new Error('Failed to created object store'));
          }
          thingsOS.transaction.onerror = (ev) => {
            reject(new Error('Failed to created object store'));
          }
          entriesOD.transaction.onerror = (ev) => {
            reject(new Error('Failed to created object store'));
          }
        } catch (error) {
          // TODO: handle unsuccessful startup better
          console.error(error);
        }
      }
    });
  }

  async getTrackBySlug(slug: string): Promise<ITrack> {
    let tx = this.DB?.transaction(OS_NAME_TRACKS, "readonly");
    return new Promise((resolve, reject) => {

      if (!tx) {
        reject(new Error('Failed to create transaction'));
        return;
      }

      let os = tx?.objectStore(OS_NAME_TRACKS);
      let index = os?.index('slug');
      let rq = index.get(slug);

      tx.oncomplete = (ev) => {
        if (DEBUG) console.log(ev, rq.result);
        resolve(rq.result as ITrack);
      }

      tx.onerror = (ev) => {
        let event = ev as Event & { target: IDBRequest };
        reject(new Error('Failed to get tracks', { cause: event.target.error }));
      };
    });
  }

  async getTracks(): Promise<ITrack[]> {
    let tx = this.DB?.transaction(OS_NAME_TRACKS, "readonly");
    return new Promise((resolve, reject) => {

      if (!tx) {
        reject(new Error('Failed to create transaction'));
        return;
      }

      let os = tx?.objectStore(OS_NAME_TRACKS);
      let rq = os.getAll();

      tx.oncomplete = (ev) => {
        resolve(rq.result);
      }

      tx.onerror = (ev) => {
        let event = ev as Event & { target: IDBRequest };
        reject(new Error('Failed to get tracks', { cause: event.target.error }));
      };
    });
  }

  async getThingById(id: ITrackThing['id']): Promise<ITrackThing> {
    let tx = this.DB?.transaction(OS_NAME_THINGS, "readonly");
    return new Promise((resolve, reject) => {

      if (!tx) {
        reject(new Error('Failed to create transaction'));
        return;
      }

      let os = tx?.objectStore(OS_NAME_THINGS);
      let index = os?.index('id');
      let rq = index.get(id);

      tx.oncomplete = (ev) => {
        if (DEBUG) console.log(ev, rq.result);
        resolve(rq.result as ITrackThing);
      }

      tx.onerror = (ev) => {
        let event = ev as Event & { target: IDBRequest };
        reject(new Error('Failed to get thing', { cause: event.target.error }));
      };
    });
  }

  async getThingsById(id: ITrack['id']): Promise<Array<ITrackThing>> {
    let tx = this.DB?.transaction(OS_NAME_THINGS, "readonly");
    return new Promise((resolve, reject) => {

      if (!tx) {
        reject(new Error('Failed to create transaction'));
        return;
      }

      let os = tx?.objectStore(OS_NAME_THINGS);
      let index = os.index('trackId');
      let rq = index.getAll(id);

      tx.oncomplete = (ev) => {
        resolve(rq.result as ITrackThing[]);
      }

      tx.onerror = (ev) => {
        let event = ev as Event & { target: IDBRequest };
        reject(new Error('Failed to get tracks', { cause: event.target.error }));
      };
    });
  }

  async getEntriesRange(thing: ITrackThing, start: Date, end: Date): Promise<ITrackEntry[]> {
    let tx = this.DB?.transaction(OS_NAME_ENTRIES, "readonly");
    return new Promise((resolve, reject) => {

      if (!tx) {
        reject(new Error('Failed to create transaction'));
        return;
      }

      let os = tx?.objectStore(OS_NAME_ENTRIES);
      let index = os.index('startAt');
      let rq = index.getAll(IDBKeyRange.bound(start, end)) as IDBRequest<ITrackEntry[]>;

      tx.oncomplete = (ev) => {
        console.log(rq.result);
        let filteredResults = rq.result.filter(entry => entry.thingId === thing.id);
        resolve(filteredResults);
      }

      tx.onerror = (ev) => {
        let event = ev as Event & { target: IDBRequest };
        reject(new Error('Failed to get tracks', { cause: event.target.error }));
      };
    });
  }

  async addTrack(track: ITrack): Promise<string> {
    let key = await this.add(OS_NAME_TRACKS, track);
    return key as string;
  }

  async addThing(thing: Omit<ITrackThing, 'id' | 'entries'>): Promise<ITrackThing> {
    let key = await this.add(OS_NAME_THINGS, thing);
    let _thing: ITrackThing = {
      ...thing,
      entries: [],
      id: key as number,
    };
    document.dispatchEvent(new CustomEvent(
      thingCreateEventType,
      { detail: _thing, bubbles: true }
    ));
    return _thing;
  }

  async addEntry(entry: Omit<ITrackEntry, 'id'>): Promise<ITrackEntry> {
    let key = await this.add(OS_NAME_ENTRIES, entry);
    let _entry: ITrackEntry = {
      ...entry,
      id: key as string,
    }
    document.dispatchEvent(new CustomEvent(
      entryCreateEventType,
      { detail: _entry, bubbles: true }
    ));
    return _entry;
  }
}