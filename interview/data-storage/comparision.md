 Q1. Scale
      - Single machine, simple ops → PostgreSQL / MySQL
      - Single machine, speed critical → Redis
      - Multi-machine, automatic sharding → Cassandra / Bigtable / DynamoDB / Spanner

  Q2. Data shape
      - Flat key → value → Redis, DynamoDB
      - Structured rows + sparse columns → Bigtable, Cassandra
      - Nested documents → MongoDB
      - Relationships → Neo4j
      - Raw files → S3 / GFS

  Q3. Access pattern
      - Random single key → Redis, DynamoDB
      - Range scan within partition → Cassandra
      - Range scan across entire keyspace → Bigtable
      - Full-text search → Elasticsearch
      - Aggregations / analytics → BigQuery, Redshift, Druid

  Q4. Write pattern
      - Write-heavy, append → Bigtable, Cassandra (LSM-tree)
      - Read-heavy, update in place → PostgreSQL (B-tree)
      - Time-series, high ingestion → InfluxDB, Druid

  Q5. CAP tradeoff
      - AP (always available, tolerate stale) → Cassandra, DynamoDB
      - CP (consistent, tolerate brief downtime) → Bigtable, Spanner, Zookeeper

  Q6. Query complexity
      - Key lookups + range scans only → Bigtable, Cassandra
      - Aggregations + filtering, no joins → Spanner, CockroachDB
      - Full SQL + joins + transactions → PostgreSQL, Spanner

  Q7. Latency requirement
      - Microseconds → Redis
      - Milliseconds → Bigtable, DynamoDB, Cassandra
      - Seconds acceptable → BigQuery, Redshift (OLAP)




This file contains the different data storage engine comarision
  ┌───────────────────────────────────┬──────────────────┬────────────────────────────────────────┐
  │             Use case              │   Pick instead   │                  Why                   │
  ├───────────────────────────────────┼──────────────────┼────────────────────────────────────────┤
  │ Need high availability always     │ Dynamo/Cassandra │ AP, survives node loss without         │
  │                                   │                  │ downtime                               │
  ├───────────────────────────────────┼──────────────────┼────────────────────────────────────────┤
  │ Rich queries, ad-hoc filtering    │ PostgreSQL       │ SQL is better when data fits one       │
  │                                   │                  │ machine                                │
  ├───────────────────────────────────┼──────────────────┼────────────────────────────────────────┤
  │ Simple key lookups, no range      │ Redis/DynamoDB   │ Bigtable's sorted model is overkill    │
  │ scans                             │                  │                                        │
  ├───────────────────────────────────┼──────────────────┼────────────────────────────────────────┤
  │ Unstructured data                 │ GFS/S3           │ Bigtable adds complexity you don't     │
  │                                   │                  │ need                                   │
  └───────────────────────────────────┴──────────────────┴────────────────────────────────────────┘