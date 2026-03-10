Object Storage — System Design Notes

---
Storage Type Comparison

Block storage:
  Abstraction: raw disk blocks; OS sees it as a hard drive
  Access: low-level, sector-by-sector
  Latency: <1ms (NVMe SSD), ~1ms (SAN)
  Use cases: VM disks, database data files, OS boot volumes
  Examples: AWS EBS, GCP Persistent Disk, Azure Managed Disks
  Properties: mountable as filesystem, random read/write, not shared across nodes by default

File storage (NFS / shared filesystem):
  Abstraction: hierarchical directory tree with filenames
  Access: POSIX filesystem operations (open, read, write, seek)
  Use cases: shared home directories, legacy apps that need a filesystem, ML training data mounts
  Examples: AWS EFS, NFS, Azure Files
  Properties: mountable by multiple nodes simultaneously, familiar API, metadata in directories

Object storage:
  Abstraction: flat namespace of objects identified by a key (no directory hierarchy, just key prefixes)
  Access: HTTP API (GET, PUT, DELETE) — not mountable as filesystem
  Latency: 10-100ms (network + distributed storage overhead)
  Durability: extreme (11 nines = 99.999999999%)
  Scalability: essentially unlimited capacity; no pre-provisioning
  Use cases: static assets, images, video, backups, data lake, log archives
  Examples: AWS S3, GCP Cloud Storage, Azure Blob Storage, MinIO (self-hosted)
  Properties: metadata per object (key-value pairs), versioning, lifecycle policies

Decision rule:
  Need random low-latency read/write → block storage
  Need POSIX filesystem shared across nodes → file storage
  Need unlimited durable cheap storage for blobs → object storage
  Need to query the data with SQL → load from object storage into a database or use Athena/BigQuery

---
S3 Architecture

Buckets:
  Top-level namespace unit; unique within a region (globally unique bucket name)
  Not a directory — just a logical grouping with access control policies
  One AWS account can have up to 100 buckets (soft limit, can be increased)

Objects:
  Immutable blob of bytes identified by a key
  Max size: 5TB per object (for single PUT: 5GB; beyond that requires multipart upload)
  Metadata: system metadata (Content-Type, ETag, Last-Modified) + user-defined key-value pairs

Keys:
  Flat string like "images/users/42/avatar.png" — looks like a path but is just a string
  "/" in keys used by S3 console to simulate folder hierarchy (not real directories)

Regions and availability zones:
  Bucket exists in one region — data not automatically replicated cross-region
  Within a region, S3 replicates data across at least 3 AZs
  11 nines durability achieved by cross-AZ redundancy + erasure coding

---
Multipart Upload

Why it exists:
  Single PUT is limited to 5GB and is unreliable for large files (one network error = restart from scratch)
  Multipart upload: split file into parts, upload in parallel, retry failed parts independently

How it works:
  1. Initiate: POST /bucket/key?uploads → returns UploadId
  2. Upload parts: PUT /bucket/key?partNumber=N&uploadId=X → each part gets an ETag
     - Minimum part size: 5MB (except last part which can be smaller)
     - Maximum parts: 10,000
  3. Complete: POST /bucket/key?uploadId=X with list of {partNumber, ETag} → S3 assembles the object
  4. Abort (optional): cleanup incomplete multipart uploads to avoid storage charges

Parallelism:
  Parts can be uploaded concurrently — limited by client concurrency and network
  Example: 1GB file split into 100 × 10MB parts, uploaded with 10 concurrent connections
  → ~10x speedup vs sequential upload

---
Pre-Signed URLs

What they are:
  A time-limited URL containing embedded credentials (AWS Signature) that grants temporary access
  to a specific S3 object, without requiring the requestor to have AWS credentials

Generation:
  App server calls AWS SDK to generate a pre-signed URL with: bucket, key, operation, expiry
  URL contains: the key, expiry timestamp, and a cryptographic signature

Use case — client-direct upload:
  Problem without pre-signed URLs: client uploads file to your server → your server uploads to S3
    → double bandwidth cost, your server is on the hot path
  With pre-signed URLs:
    1. Client requests an upload URL from your API
    2. API generates pre-signed PUT URL, returns it to client (signed for 15 min)
    3. Client PUTs file directly to S3 using the pre-signed URL
    4. Client notifies your API that upload is complete
    → your servers never touch the file bytes; S3 absorbs the ingress

Expiry:
  Maximum validity: 7 days (168 hours) for presigned URLs generated with IAM user credentials
  Shorter is better — principle of least privilege applied to time
  Pre-signed URL for a private object expires → request returns 403 Forbidden

Security implications:
  Pre-signed URL is effectively a bearer token for that specific operation
  Anyone with the URL can perform the operation until expiry — do not log or share these URLs
  If a pre-signed URL leaks, you cannot revoke it before expiry (no blocklist mechanism)
  Mitigation: very short expiry windows (minutes, not hours) for sensitive content

---
S3 Consistency Model

Before November 2020:
  PUT for a new key: eventually consistent (GET after PUT might return 404 briefly)
  PUT overwrite: eventually consistent (GET might return old version)
  DELETE: eventually consistent

After November 2020 (current):
  Strong read-after-write consistency for all operations — GET, PUT, DELETE, LIST
  A successful PUT is immediately visible to all subsequent GETs from anywhere
  LIST immediately reflects newly added or deleted objects

What changed: AWS rebuilt the S3 metadata subsystem to provide strong consistency
  → no application-level workarounds needed (previously: add version to key, read with delay)

Implication for system design:
  You can now rely on PUT→GET consistency without extra coordination
  LIST operations are also strongly consistent (previously lagged)

---
Storage Classes

S3 Standard:
  Durability: 11 nines; Availability: 99.99%
  Retrieval: milliseconds
  Cost: highest storage cost per GB
  Use when: frequently accessed data (web assets, active application data)

S3 Standard-IA (Infrequent Access):
  Same durability and availability as Standard
  Retrieval: milliseconds (same as Standard)
  Cost: lower storage cost per GB, but per-retrieval fee + 30-day minimum storage charge
  Use when: data accessed ~once per month (DR copies, older logs, thumbnails rarely viewed)

S3 One Zone-IA:
  Data stored in one AZ only (lower redundancy)
  Cost: 20% cheaper than Standard-IA
  Use when: reproducible data where loss of one AZ is acceptable (re-creatable thumbnails, derived data)

S3 Glacier Instant Retrieval:
  Same millisecond retrieval as Standard-IA
  Lowest cost for infrequently accessed data with instant retrieval requirement
  90-day minimum storage duration

S3 Glacier Flexible Retrieval (formerly S3 Glacier):
  Retrieval: minutes to hours (expedited: 1-5 min, standard: 3-5 hr, bulk: 5-12 hr)
  Significantly lower cost — designed for archival
  Use when: compliance archives, long-term backups, data retained for years

S3 Glacier Deep Archive:
  Retrieval: 12-48 hours
  Cheapest storage class
  Use when: 7-10 year compliance retention, true cold archive

S3 Intelligent-Tiering:
  Automatically moves objects between frequent/infrequent access tiers based on access patterns
  Small monthly monitoring fee per object
  Use when: access patterns are unknown or unpredictable

---
Durability vs Availability

Durability (11 nines = 99.999999999%):
  Probability that a stored object will NOT be lost or corrupted
  11 nines = expected loss of 1 object per 100 billion object-years
  Achieved by: erasure coding + replication across 3+ AZs
  This is about data surviving — not about whether you can access it right now

Availability (99.99% for S3 Standard):
  Percentage of time S3 correctly responds to requests
  99.99% = ~52 minutes of downtime per year
  A service can be unavailable (temporarily) while still being durable (data is not lost)

The key distinction:
  Durability problem: "Is the data still there?" — S3 almost never loses data
  Availability problem: "Can I access it right now?" — S3 has rare brief outages
  You can have high durability with low availability (data intact but service is down)
  You cannot recover from low durability with high availability (data is gone)

---
CDN Integration

CloudFront in front of S3:
  Requests route to nearest edge location (100+ PoPs globally)
  Cache hit: served from edge — ~5-10ms latency, no S3 egress cost
  Cache miss: edge fetches from S3 origin, caches it at the edge
  Benefit: reduced latency for global users, lower S3 egress costs (CloudFront egress is cheaper)

Cache invalidation:
  CreateInvalidation API: invalidate specific paths or wildcard (/images/*)
  Propagates to all edge locations in ~15 minutes
  Cost: first 1000 invalidations/month free, then $0.005 per path
  Alternative: versioned URLs (image_v3.png) — no invalidation needed, old version just gets evicted

When to use:
  Static assets (JS, CSS, images): always put behind CDN
  Dynamic API responses: CDN can cache if Cache-Control headers are set correctly (short TTL)
  Private S3 content: use CloudFront signed URLs (not S3 pre-signed) for access-controlled distribution

---
Access Control

Three mechanisms (from most specific to most broad):

ACLs (Access Control Lists):
  Object-level or bucket-level; legacy mechanism; AWS recommends disabling in new accounts
  Canned ACLs: private, public-read, public-read-write
  Avoid: coarse-grained, hard to audit, replaced by policies

Bucket policies (JSON resource-based policies):
  Attached to a bucket; can grant/deny access to any AWS principal or public
  Example: "Allow any unauthenticated user to GET objects in /public/*"
  Use for: cross-account access, public bucket hosting, enforcing HTTPS only (aws:SecureTransport)

IAM policies (identity-based):
  Attached to IAM users, roles, or groups
  Grants permissions to S3 actions from a specific identity
  Use for: EC2 instances with instance roles, Lambda functions, app deployments

Evaluation: request is allowed only if at least one policy allows AND no explicit deny exists
Block Public Access: account/bucket-level override that blocks all public access regardless of policies

---
Object Storage vs Database for Binary Data

Store in object storage (S3) when:
  - File is large (>100KB) — databases degrade significantly with large BLOBs
  - Access pattern is read-once-write-rarely (backups, media files, archives)
  - Need public/CDN-accessible URLs
  - Storage cost matters — S3 is 10-100x cheaper per GB than RDS
  - Files are unstructured and don't need transactional writes with other data

Store in database when:
  - File is tiny (<10KB) and closely related to a row (small avatar, signature image)
  - Need ACID transactions that span both the binary data and its metadata
  - Convenience outweighs cost and performance at low scale

Hybrid pattern (most common):
  Store the file in S3 → store the S3 key (string) in the database row
  Database row: {user_id: 42, avatar_key: "avatars/user-42-v3.jpg", updated_at: ...}
  → transactional integrity for metadata, cost efficiency and CDN-ability for the file
