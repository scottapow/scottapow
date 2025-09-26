import type { ITrack } from "../../track/types";
import type { IStore } from "./types";

const V = 1; // cannot be a float

export class ClientDB {
  DB?: IDBDatabase;
  state: 'ready' | 'error' | 'idle' = 'idle';
  private OS_NAME_TRACKS = 'tracks' as const;
  private OS_NAME_THINGS = 'things' as const;

  async connect(name: string): Promise<typeof this> {
    return new Promise((resolve, reject) => {
      console.log('hello');
      const request = window.indexedDB.open(name, V);

      request.onerror = () => {
        console.log('onerror');
        if (request.error) {
          console.log('hello');
          reject(request.error.message);
        } else {
          reject(new Error('Unknown error on IndexedDB request'));
        }
      }

      request.onsuccess = (ev) => {
        console.log('onsuccess');
        this.DB = (ev.target as EventTarget & { result: IDBDatabase })?.result;
        resolve(this);
      }

      request.onupgradeneeded = (ev) => {
        console.log('onupgradeneeded');
        this.DB = (ev.target as EventTarget & { result: IDBDatabase })?.result;
        try {
          const tracksOS: IDBObjectStore = this.DB?.createObjectStore(this.OS_NAME_TRACKS, { keyPath: 'id' })!;
          tracksOS?.createIndex('id', 'id', { unique: true }); // PK

          const thingsOS: IDBObjectStore = this.DB?.createObjectStore(this.OS_NAME_THINGS, { keyPath: 'id' })!;
          thingsOS?.createIndex('id', 'id', { unique: true }); // PK
          thingsOS?.createIndex('trackId', 'trackId', { unique: false }); // FK

          tracksOS.transaction.onerror = (ev) => {
            reject(new Error('Failed to created object store'));
          }
          thingsOS.transaction.onerror = (ev) => {
            reject(new Error('Failed to created object store'));
          }
        } catch (error) {
          // TODO: handle unsuccessful startup better
          console.error(error);
        }
      }
    });
  }

  async getTracks(): Promise<ITrack[]> {
    let tx = this.DB?.transaction(this.OS_NAME_TRACKS, "readonly");
    return new Promise((resolve, reject) => {

      if (!tx) {
        reject(new Error('Failed to create transaction'));
        return;
      }

      let os = tx?.objectStore(this.OS_NAME_TRACKS);
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

  async addTrack(track: ITrack): Promise<string> {
    console.log('add track');
    let tx = this.DB?.transaction(this.OS_NAME_TRACKS, "readwrite");
    return new Promise((resolve, reject) => {

      if (!tx) {
        reject(new Error('Failed to create transaction'));
        return;
      }

      let os = tx?.objectStore(this.OS_NAME_TRACKS);
      let rq = os?.add(track);

      tx.oncomplete = (ev) => {
        // console.log(ev);
        // let event = ev as Event & { target: IDBRequest<string> };
        resolve(rq.result as string);
      }

      tx.onerror = (ev) => {
        let event = ev as Event & { target: IDBRequest };
        reject(new Error('Failed to add track', { cause: event.target.error }));
      };
    });
  }
}