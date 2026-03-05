  Bigtable Reading Guide

  Paper: Bigtable: A Distributed Storage System for Structured Data (Google, 2006)
  URL: https://static.googleusercontent.com/media/research.google.com/en//archive/bigtable-osdi06.pdf

  ---
  What Bigtable Is — Before You Read

  Bigtable is Google's internal large-scale structured storage system. It sits on the CP side of CAP — the direct contrast to Dynamo. Where Dynamo sacrifices
  consistency for availability, Bigtable sacrifices availability for strong consistency and structured data.

  It's the foundation for: HBase, Cassandra's data model, DynamoDB's structured types, Google Cloud Bigtable.

  ---
  Reading Order

  §1 — Introduction
  - Why Google built it, what problems it solves
  - What "structured data at Google scale" means

  §2 — Data Model ← Most important section
  - The wide-column model: rows, column families, columns, timestamps
  - Spend the most time here — this is what every interview question about Bigtable is actually about

  §4 — Building Blocks
  - SSTable format (how data is stored on disk)
  - Chubby (distributed lock service — how master election and metadata works)
  - GFS (underlying file system — just know it exists)

  §5 — Implementation
  - Tablet servers — how reads and writes actually happen
  - Master server — what it does and crucially what it does NOT do
  - Compactions — minor, merging, major

  §6 — Refinements
  - Locality groups — performance optimization
  - Bloom filters — read optimization
  - Caching — two levels

  ---
  Sections to Skip

  - §3 (API) — skim, 1 minute
  - §7 (Performance numbers) — skim the summary only
  - §8 (Real applications) — skip

  ---
  Key Questions to Answer While Reading

  1. What are rows, column families, columns, and timestamps — and how do they relate?
  2. What is an SSTable and how does a read find data across multiple SSTables?
  3. What does the master server do — and what does it NOT handle?
  4. What is a tablet and how does Bigtable scale horizontally?
  5. What is Chubby and why does Bigtable need it?

  ---
  Retention Questions When You Return

  1. A row in Bigtable can have thousands of columns, and two rows don't need the same columns. How is this possible and what problem does it solve?
  2. Bigtable is CP, Dynamo is AP. What does that mean concretely for how each handles a node failure?
  3. The master server is not on the read/write path. How does a client find the right tablet server then?