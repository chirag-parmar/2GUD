```bash
cd node
docker build --tag zama-node .
cd ../client
docker build --tag zama-client .
cd ..
docker compose up -d
docker exec -it zama-client-1 /bin/sh
```