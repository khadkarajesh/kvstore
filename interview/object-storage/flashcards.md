## Object Storage — Flashcards

Format: Q → A
Cover the answer, recall it, then check.

---

**Q: What are the three storage types and their core access model?**
A: Block storage: raw disk blocks, mountable as filesystem, low latency (<1ms), used for VM disks and DBs. File storage: hierarchical POSIX filesystem, mountable by multiple nodes (NFS/EFS), shared access. Object storage: flat key-value namespace over HTTP (GET/PUT/DELETE), not mountable, designed for unlimited durable blob storage.

---

**Q: Why can't you mount S3 as a filesystem for a production database?**
A: S3 access latency is 10-100ms (network round trip to distributed storage). Databases require <1ms random read/write latency. S3 also does not support POSIX semantics (no append-in-place, no locking, no atomic rename). Use block storage (EBS) for database data files.

---

**Q: What is an S3 key and is "images/users/42/avatar.png" a real directory path?**
A: An S3 key is a flat string — there are no real directories in S3. "images/users/42/avatar.png" is a single string key. The S3 console simulates folder hierarchy by treating "/" as a delimiter, but the underlying storage model is a flat namespace.

---

**Q: How does S3 achieve 11 nines of durability?**
A: Data is automatically replicated across at least 3 Availability Zones within the region, combined with erasure coding. An AZ-level failure loses no data. 11 nines means an expected loss of 1 object per 100 billion object-years.

---

**Q: What is the difference between S3 durability (11 nines) and S3 availability (99.99%)?**
A: Durability: the data will not be lost or corrupted — this is about the survival of stored objects (99.999999999%). Availability: the service will respond to requests — this is about uptime (99.99% ≈ 52 min/year downtime). High durability with temporary unavailability is normal. Low durability means data is gone permanently.

---

**Q: What triggers the need for multipart upload and what is the minimum part size?**
A: A single PUT is limited to 5GB. Beyond that, multipart upload is required. Multipart is also beneficial for large files at any size (parallel uploads, retry individual parts). Minimum part size: 5MB (except the final part). Maximum parts: 10,000.

---

**Q: Walk through the three steps of a multipart upload.**
A: 1) Initiate — POST ?uploads, receive UploadId. 2) Upload parts — PUT each part with partNumber + UploadId, receive ETag per part. Parts can be uploaded in parallel. 3) Complete — POST with UploadId + list of {partNumber, ETag}, S3 assembles the final object.

---

**Q: What is a pre-signed URL and what problem does it solve?**
A: A time-limited URL with embedded AWS credentials (signature) that grants access to a specific S3 object. Solves the double-bandwidth problem: instead of client → server → S3, the client uploads/downloads directly to/from S3. The server generates the URL but never touches the file bytes.

---

**Q: What is the security risk of a pre-signed URL and how do you mitigate it?**
A: Pre-signed URLs are bearer tokens — anyone who has the URL can use it until expiry. They cannot be revoked before expiry. Mitigation: use the shortest expiry that is practical (minutes, not hours/days); never log these URLs; treat them like temporary passwords.

---

**Q: What changed in S3's consistency model in November 2020?**
A: S3 moved from eventual consistency to strong read-after-write consistency for all operations. Before: a GET after a PUT might return 404 or the old version. After: a successful PUT is immediately visible to all subsequent GETs and LISTs from anywhere. No more application-level workarounds needed.

---

**Q: S3 Standard vs S3 Standard-IA — when do you use Standard-IA?**
A: Standard-IA (Infrequent Access) is for data accessed roughly once a month or less. It has the same latency and durability as Standard, but lower storage cost per GB offset by a per-retrieval fee + 30-day minimum storage charge. Using it for frequently accessed data actually costs more than Standard.

---

**Q: S3 Glacier Flexible Retrieval vs Glacier Deep Archive — what is the difference?**
A: Glacier Flexible Retrieval: retrieval in minutes to hours (expedited: 1-5 min, standard: 3-5 hr). Glacier Deep Archive: retrieval in 12-48 hours. Deep Archive is the cheapest S3 storage class, designed for 7-10 year compliance retention where retrieval is extremely rare.

---

**Q: What is S3 Intelligent-Tiering and when is it appropriate?**
A: Automatically moves objects between frequent and infrequent access tiers based on actual access patterns. Charges a small monthly monitoring fee per object. Appropriate when access patterns are unpredictable or mixed — you pay for optimization instead of manually choosing the right tier.

---

**Q: Why put CloudFront in front of S3 and what does cache invalidation cost?**
A: CloudFront caches S3 objects at 100+ edge locations globally, reducing latency to ~5-10ms and lowering S3 egress costs (CloudFront egress is cheaper than direct S3). Cache invalidation: first 1,000 paths/month are free; $0.005 per path after that. Alternative: versioned keys (image_v3.png) avoid invalidation entirely.

---

**Q: What are the three S3 access control mechanisms and which should you prefer?**
A: ACLs (legacy, object/bucket level — avoid in new accounts), bucket policies (JSON resource-based, good for cross-account and public access), IAM policies (identity-based, attached to roles/users). Prefer bucket policies + IAM policies over ACLs. AWS now recommends disabling ACLs.

---

**Q: A user uploads a profile image. You store it in S3 and save the S3 URL in your database. What is the better pattern and why?**
A: Store the S3 key (not the full URL) in the database. The URL format includes the region and bucket name which may change. The key is stable. Construct the URL at read time from the key + known bucket. Also enables easy CDN URL generation by swapping the hostname.

---

**Q: When should you store binary data in a database instead of S3?**
A: Only when files are tiny (<10KB) and require transactional writes with related row data. For everything else, use S3 with the key stored in the database (hybrid pattern). Databases degrade with large BLOBs; S3 is 10-100x cheaper per GB and CDN-ready.

---

**Q: What is the max object size in S3 and how do you upload something larger than the single-PUT limit?**
A: Max object size: 5TB. Single PUT limit: 5GB. For objects larger than 5GB (or any large object for reliability), use multipart upload — split into parts, upload in parallel, complete with a manifest of part ETags.

---

**Q: A pre-signed URL was accidentally committed to a public GitHub repo. The URL expires in 6 hours. What do you do?**
A: You cannot revoke the pre-signed URL itself. Immediate mitigation: delete or rename the S3 object (changing the key invalidates any existing pre-signed URLs for the old key). If deletion is not possible, assess the sensitivity of the exposed object and consider blocking access via bucket policy. Rotate any long-lived secrets that were used to generate the URL.

---

**Q: S3 bucket in us-east-1 stores your application data. Your app needs to serve users in Europe with low latency. What are your options?**
A: 1) CloudFront distribution with S3 as origin — caches at European edge locations, sub-10ms for cached content. 2) S3 Cross-Region Replication (CRR) — async replication to an eu-west-1 bucket, then serve from the nearest region. 3) Multi-region active-active — serve reads from nearest region, handle conflict resolution. For static content, CloudFront is usually sufficient.

---
