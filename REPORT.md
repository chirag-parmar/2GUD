### Selecting an appropriate tree type

We can use the usual merkle trees, but the problem is that they need a lot of hashes to be computed. If we increase the radix of these trees such that each parent node is a hash of `n` child nodes (`n`-ary merkle tree) then we save some hashing computations but increase our proof size. 

|- Type -|- no. of leaves -|- proof size -|- hashes to compute (in total) -|- hashes to compute (for adding a new file) -|- hashes (for updating a file) -|
| Merkle Tree | n | O(log_2(n)) | n + n/2 + n/4 + .... + 1 = 2n(1+n) => O(n^2) | log_2(n) | log_2(n) |
| k-width Merkle Tree | n | log_K(n) | n + n/k + n/(k^2) + n/(k^3) + ... + 1 = kn(1+n)/(k - 1) =

### Server assumptions

* The files are small enough for HTTP multipart form data to ignore usage of FTP.
* The files are larger than what non multipart HTTP requests can handle.
* max file size is 1MB, following the above two rules.