export type WrappedIDBObjectStore<N extends string, T extends {}> =
  Pick<ReturnType<IDBTransaction['objectStore']>, 'autoIncrement' | 'indexNames' | 'keyPath' | 'transaction'> &
  {
    /** The **`name`** property of the IDBObjectStore interface indicates the name of this object store.
     *
     * [MDN Reference](https://developer.mozilla.org/docs/Web/API/IDBObjectStore/name) */
    name: N;
    /** The **`transaction`** read-only property of the object store belongs.
     *
     * [MDN Reference](https://developer.mozilla.org/docs/Web/API/IDBObjectStore/transaction) */
    readonly transaction: IDBTransaction;
    /** The **`add()`** method of the IDBObjectStore interface returns an IDBRequest object, and, in a separate thread, creates a structured clone of the value, and stores the cloned value in the object store.
     *
     * [MDN Reference](https://developer.mozilla.org/docs/Web/API/IDBObjectStore/add) */
    add(value: T, key?: IDBValidKey): IDBRequest<IDBValidKey>;
    /** The **`clear()`** method of the IDBObjectStore interface creates and immediately returns an IDBRequest object, and clears this object store in a separate thread.
     *
     * [MDN Reference](https://developer.mozilla.org/docs/Web/API/IDBObjectStore/clear) */
    clear(): IDBRequest<undefined>;
    /** The **`count()`** method of the IDBObjectStore interface returns an IDBRequest object, and, in a separate thread, returns the total number of records that match the provided key or of records in the store.
     *
     * [MDN Reference](https://developer.mozilla.org/docs/Web/API/IDBObjectStore/count) */
    count(query?: IDBValidKey | IDBKeyRange): IDBRequest<number>;
    /** The **`createIndex()`** method of the field/column defining a new data point for each database record to contain.
     *
     * [MDN Reference](https://developer.mozilla.org/docs/Web/API/IDBObjectStore/createIndex) */
    createIndex(name: keyof T, keyPath: keyof T | (keyof T)[], options?: IDBIndexParameters): IDBIndex;
    /** The **`delete()`** method of the and, in a separate thread, deletes the specified record or records.
     *
     * [MDN Reference](https://developer.mozilla.org/docs/Web/API/IDBObjectStore/delete) */
    delete(query: IDBValidKey | IDBKeyRange): IDBRequest<undefined>;
    /** The **`deleteIndex()`** method of the the connected database, used during a version upgrade.
     *
     * [MDN Reference](https://developer.mozilla.org/docs/Web/API/IDBObjectStore/deleteIndex) */
    deleteIndex(name: keyof T): void;
    /** The **`get()`** method of the IDBObjectStore interface returns an IDBRequest object, and, in a separate thread, returns the object selected by the specified key.
     *
     * [MDN Reference](https://developer.mozilla.org/docs/Web/API/IDBObjectStore/get) */
    get(query: IDBValidKey | IDBKeyRange): IDBRequest<any>;
    /** The **`getAll()`** method of the containing all objects in the object store matching the specified parameter or all objects in the store if no parameters are given.
     *
     * [MDN Reference](https://developer.mozilla.org/docs/Web/API/IDBObjectStore/getAll) */
    getAll(query?: IDBValidKey | IDBKeyRange | null, count?: number): IDBRequest<any[]>;
    /** The `getAllKeys()` method of the IDBObjectStore interface returns an IDBRequest object retrieves record keys for all objects in the object store matching the specified parameter or all objects in the store if no parameters are given.
     *
     * [MDN Reference](https://developer.mozilla.org/docs/Web/API/IDBObjectStore/getAllKeys) */
    getAllKeys(query?: IDBValidKey | IDBKeyRange | null, count?: number): IDBRequest<IDBValidKey[]>;
    /** The **`getKey()`** method of the and, in a separate thread, returns the key selected by the specified query.
     *
     * [MDN Reference](https://developer.mozilla.org/docs/Web/API/IDBObjectStore/getKey) */
    getKey(query: IDBValidKey | IDBKeyRange): IDBRequest<IDBValidKey | undefined>;
    /** The **`index()`** method of the IDBObjectStore interface opens a named index in the current object store, after which it can be used to, for example, return a series of records sorted by that index using a cursor.
     *
     * [MDN Reference](https://developer.mozilla.org/docs/Web/API/IDBObjectStore/index) */
    index(name: keyof T): IDBIndex;
    /** The **`put()`** method of the IDBObjectStore interface updates a given record in a database, or inserts a new record if the given item does not already exist.
     *
     * [MDN Reference](https://developer.mozilla.org/docs/Web/API/IDBObjectStore/put) */
    put(value: T, key?: IDBValidKey): IDBRequest<IDBValidKey>;
  }


export type Success = boolean;

export interface IStore {
  DB: IDBDatabase | null;
  init: () => Promise<Success>;
}