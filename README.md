### Setup nodes

```bash
cd node
docker build --tag zama-node .
cd ../client/uploadables
python ./create_dummies.py
cd ..
docker build --tag zama-client .
cd uploadables
cd ..
docker compose up -d
docker exec -it zama-client-1 /bin/sh
```

### Client operations

upload dummy files to the servers
```bash
./client --upload=true
```

use the first merkle hash from the above printed log and run the query beloe to download the file
```bash
./client --merkle=<FIRST MERKLE HASH FROM PREVIOUS STEP> --ip 172.10.0.2 --index 12
```

kill 1st node and then alternate IPs from 172.10.0.5-7 (one of them must be the replicas turned primary)
```bash
./client --merkle=60e37672cd54ac0f3d1e9e3e53e398e62d468860843dba0eabca8e8e510e9b57 --ip <IP> --index 12
```