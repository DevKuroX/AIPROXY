const SQLiteDisabledError = new Error(
  "SQLite has been removed. Use the Go backend API instead."
);

export function createSqljsAdapter() { throw SQLiteDisabledError; }
export function createBetterSqliteAdapter() { throw SQLiteDisabledError; }
export function createBunSqliteAdapter() { throw SQLiteDisabledError; }
export function createNodeSqliteAdapter() { throw SQLiteDisabledError; }
export default { createSqljsAdapter, createBetterSqliteAdapter, createBunSqliteAdapter, createNodeSqliteAdapter };
