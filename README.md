### Setup nodes

```bash
cd node
docker build --tag zama-node .
cd ../client/uploadables
python ./create_dummies.py
cd ..
docker build --tag zama-client .
docker compose up -d
docker exec -it zama-client-1 /bin/sh
```

### Client operations

The above command will start a shell inside the client node for you, then use the below operations to interact with other nodes.

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
./client --merkle=<FIRST MERKLE HASH FROM PREVIOUS STEP> --ip <IP> --index 12
```