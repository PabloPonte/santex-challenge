## Notes and Assumptions

* A feature-flag administration application will be built.
* Information will be stored in a relational database.
* Because reads will be frequent while writes will not, an in-memory mirror of the database (possibly Redis) will be created as a cache to improve query performance.
* The cache will be updated whenever the database changes to keep both stores consistent.
* Manual cache-refresh mechanisms will be available.
* The following data is expected to be defined:
  * feature name (unique identifier)
  * feature description
  * feature status (`open`, `closed`, `whitelisted`)
  * status date (timestamp of the latest status change)
  * user whitelist (users allowed to use the feature)
* CRUD management endpoints for features (the internal API) will be developed, along with the manual cache-refresh mechanism.
* An endpoint will be developed to query the status of a specific feature for a specific user. It will read from the cache (the external API).
* A small application for managing features will be developed (backoffice).
* Security will not be implemented within this scope. The backoffice is assumed to run inside a trusted network, and the external API is assumed to be consumed by an already-secured BFF.
