### Selecting an appropriate tree type

We can use the usual merkle trees, but the problem is that they need a lot of hashes to be computed. If we increase the radix of these trees such that each parent node is a hash of `n` child nodes (`n`-ary merkle tree) then we save some hashing computations but increase our proof size. 

|- Type -|- no. of leaves -|- proof size -|- hashes to compute (in total) -|- hashes to compute (for adding a new file) -|- hashes (for updating a file) -|
| Merkle Tree | n | O(log_2(n)) | n + n/2 + n/4 + .... + 1 = 2n(1+n) => O(n^2) | log_2(n) | log_2(n) |
| k-width Merkle Tree | n | log_K(n) | n + n/k + n/(k^2) + n/(k^3) + ... + 1 = kn(1+n)/(k - 1) =

### Server assumptions

* The files are small enough for HTTP multipart form data to ignore usage of FTP.
* The files are larger than what non multipart HTTP requests can handle.
* max file size is 1MB, following the above two rules.
* server capaity minimum of 1TB = 10^6 files
* one merkle node consumes = 8 + 8 + 8 + 32 = 56 bytes (roughly) (left pointer + right pointer + weight + hash)
* one entry in the hash table consumes about 40bytes (32 bytes hash + 8 bytes int)
* total space 56*(2n - 1) + 40n = 152n - 56 ~ 152n
* if we have 1TB of space we can store 1M files = 1M merkle leaf nodes => 152 * 10^6 bytes about 152MB for one server
* A machine with 16GB of RAM can probably manage merkle trees for 100 servers at a time.
* we will take this as a limit for distributing and define a single leader distributed system for file storage
* 




1. control over indexes should be with client and not the server
2. AppendLeaf was designed so that new files can be added to the same merkle root. Good for streamed, pause and resume uploading etc.